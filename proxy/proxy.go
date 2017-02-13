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
	"github.com/wgliang/pgproxy/parser"
)

var (
	connid = uint64(0) // Self-increasing ConnectID.
)

// Start proxy server needed receive  and proxyHost, all
// the request or database's sql of receive will redirect
// to remoteHost.
func Start(proxyHost, remoteHost string, powerCallback parser.Callback) {
	defer glog.Flush()
	glog.Infof("Proxying from %v to %v\n", proxyHost, remoteHost)

	proxyAddr := getResolvedAddresses(proxyHost)
	remoteAddr := getResolvedAddresses(remoteHost)
	listener := getListener(proxyAddr)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.Errorf("Failed to accept connection '%s'\n", err)
			continue
		}
		connid++

		p := &Proxy{
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

// ResolvedAddresses of host.
func getResolvedAddresses(host string) *net.TCPAddr {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		glog.Fatalln("ResolveTCPAddr of host:", err)
	}
	return addr
}

// Listener of a net.TCPAddr.
func getListener(addr *net.TCPAddr) *net.TCPListener {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		glog.Fatalf("ListenTCP of %s error:%v", addr, err)
	}
	return listener
}

// Proxy - Manages a Proxy connection, piping data between proxy and remote.
type Proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  *net.TCPConn
	erred         bool
	errsig        chan bool
	prefix        string
}

// New - Create a new Proxy instance. Takes over local connection passed in,
// and closes it when finished.
func New(conn *net.TCPConn, proxyAddr, remoteAddr *net.TCPAddr, connid int64) *Proxy {
	return &Proxy{
		lconn:  conn,
		laddr:  proxyAddr,
		raddr:  remoteAddr,
		erred:  false,
		errsig: make(chan bool),
		prefix: fmt.Sprintf("Connection #%03d ", connid),
	}
}

// proxy.err
func (p *Proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		glog.Fatalf(p.prefix+s, err)
	}
	p.errsig <- true
	p.erred = true
}

// Proxy.start open connection to remote and start proxying data.
func (p *Proxy) start(powerCallback parser.Callback) {
	defer p.lconn.Close()
	// connect to remote server
	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		p.err("Remote connection failed: %s", err)
		return
	}
	p.rconn = rconn
	defer p.rconn.Close()
	// proxying data
	go p.pipe(p.lconn, p.rconn, powerCallback)
	go p.pipe(p.rconn, p.lconn, nil)
	// wait for close...
	<-p.errsig
}

// Proxy.pipe
func (p *Proxy) pipe(src, dst *net.TCPConn, powerCallback parser.Callback) {
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
			parser.Filter(b)
		}
		// show output
		if islocal {
			b = getModifiedBuffer(b, powerCallback)
			n, err = dst.Write(b)
		} else {
			// write out result
			n, err = dst.Write(b)
		}
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
	}
}

// ModifiedBuffer when is local and will call powerCallback function
func getModifiedBuffer(buffer []byte, powerCallback parser.Callback) []byte {
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
