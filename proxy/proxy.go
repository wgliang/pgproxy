// Copyright 2017 wgliang. All rights reserved.
// Use of this source code is governed by Apache
// license that can be found in the LICENSE file.

// Package proxy provides proxy service and redirects requests
// form proxy.Addr to remote.Addr.
package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/wgliang/pgproxy/filter"
)

var (
	connid = uint64(0) // Self-increasing ConnectID.
)

// Start proxy server needed receive  and proxyHost, all
// the request or database's sql of receive will redirect
// to remoteHost.
func Start(proxyHost, remoteHost string, powerCallback filter.Callback) {
	glog.Infof("Proxying from %v to %v\n", proxyHost, remoteHost)

	proxyAddr, remoteAddr := getResolvedAddresses(proxyHost, remoteHost)
	listener := getListener(proxyAddr)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.Errorf("Failed to accept connection '%s'\n", err)
			continue
		}
		connid++

		p := &proxy{
			lconn:  conn,
			laddr:  proxyAddr,
			raddr:  remoteAddr,
			erred:  false,
			errsig: make(chan bool),
			prefix: fmt.Sprintf("Connection #%03d ", connid),
		}
		go p.start(powerCallback)
	}
}

// ResolvedAddresses of proxyHost and remoteHost.
func getResolvedAddresses(proxyHost, remoteHost string) (*net.TCPAddr, *net.TCPAddr) {
	paddr, err := net.ResolveTCPAddr("tcp", proxyHost)
	if err != nil {
		glog.Fatalln("ResolveTCPAddr of proxyHost:", err)
	}
	raddr, err := net.ResolveTCPAddr("tcp", remoteHost)
	if err != nil {
		glog.Fatalln("ResolveTCPAddr of remoteHost:", err)
	}
	return paddr, raddr
}

// Listener of a net.TCPAddr.
func getListener(addr *net.TCPAddr) *net.TCPListener {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		glog.Fatalf("ListenTCP of %s error:%v", addr, err)
	}
	return listener
}

// proxy struct.
type proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  *net.TCPConn
	erred         bool
	errsig        chan bool
	prefix        string
}

// proxy.err
func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		glog.Fatalf(p.prefix+s, err)
	}
	p.errsig <- true
	p.erred = true
}

// proxy.start
func (p *proxy) start(powerCallback filter.Callback) {
	defer p.lconn.Close()
	// connect to remote
	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		p.err("Remote connection failed: %s", err)
		return
	}
	p.rconn = rconn
	defer p.rconn.Close()
	// bidirectional copy
	go p.pipe(p.lconn, p.rconn, powerCallback)
	go p.pipe(p.rconn, p.lconn, nil)
	// wait for close...
	<-p.errsig
}

// proxy.pipe
func (p *proxy) pipe(src, dst *net.TCPConn, powerCallback filter.Callback) {
	// data direction
	islocal := src == p.lconn
	// directional copy (64k buffer)
	buff := make([]byte, 0xffff)

	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}

		b := buff[:n]

		if string(b[0]) == "Q" {
			b = filter.ReWriteSQL(b)
		}

		// show output
		if islocal {
			b = getModifiedBuffer(b, powerCallback)
			n, err = dst.Write(b)
		} else {
			//write out result
			n, err = dst.Write(b)
		}
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
	}
}

// ModifiedBuffer when is local and will call powerCallback function
func getModifiedBuffer(buffer []byte, powerCallback filter.Callback) []byte {
	if powerCallback == nil || len(buffer) < 1 || string(buffer[0]) != "Q" || string(buffer[5:11]) != "power:" {
		return buffer
	}
	query := powerCallback(string(buffer[5:]))
	return makeMessage(query)
}

// make the query message.
func makeMessage(query string) []byte {
	queryArray := make([]byte, 0, 6+len(query))
	queryArray = append(queryArray, 'Q', 0, 0, 0, 0)
	queryArray = append(queryArray, query...)
	queryArray = append(queryArray, 0)
	binary.BigEndian.PutUint32(queryArray[1:], uint32(len(queryArray)-1))
	return queryArray
}
