// Copyright 2017 wgliang. All rights reserved.
// Use of this source code is governed by Apache
// license that can be found in the LICENSE file.

// Package proxy provides proxy service and redirects requests
// form proxy.Addr to remote.Addr.
package proxy

import (
	// "bytes"
	// "encoding/binary"
	"errors"
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
func Start(proxyHost, remoteHost string, filterCallback, returnCallBack parser.Callback) {
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
			connId: connid,
		}
		go p.service(filterCallback, returnCallBack)
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
	connId        uint64
}

// New - Create a new Proxy instance. Takes over local connection passed in,
// and closes it when finished.
func New(conn *net.TCPConn, proxyAddr, remoteAddr *net.TCPAddr, connid uint64) *Proxy {
	return &Proxy{
		lconn:  conn,
		laddr:  proxyAddr,
		raddr:  remoteAddr,
		erred:  false,
		errsig: make(chan bool),
		prefix: fmt.Sprintf("Connection #%03d ", connid),
		connId: connid,
	}
}

// proxy.err
func (p *Proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		glog.Errorf(p.prefix+s, err)
	}
	p.errsig <- true
	p.erred = true
}

// Proxy.service open connection to remote and service proxying data.
func (p *Proxy) service(filterCallback, returnCallBack parser.Callback) {
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
	go p.handleIncomingConnection(p.lconn, p.rconn, filterCallback)
	go p.handleResponseConnection(p.rconn, p.lconn, returnCallBack)
	// wait for close...
	<-p.errsig
}

// Proxy.handleIncomingConnection
func (p *Proxy) handleIncomingConnection(src, dst *net.TCPConn, Callback parser.Callback) {
	// directional copy (64k buffer)
	buff := make([]byte, 0xffff)

	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b, err := getModifiedBuffer(buff[:n], Callback)
		if err != nil {
			p.err("%s\n", err)
			err = dst.Close()
			if err != nil {
				glog.Errorln(err)
			}
			return
		}

		n, err = dst.Write(b)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
	}
}

// Proxy.handleResponseConnection
func (p *Proxy) handleResponseConnection(src, dst *net.TCPConn, Callback parser.Callback) {
	// directional copy (64k buffer)
	buff := make([]byte, 0xffff)

	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b := setResponseBuffer(p.erred, buff[:n], Callback)

		n, err = dst.Write(b)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
	}
}

// ModifiedBuffer when is local and will call filterCallback function
func getModifiedBuffer(buffer []byte, filterCallback parser.Callback) (b []byte, err error) {
	if len(buffer) > 0 && string(buffer[0]) == "Q" {
		if !filterCallback(buffer) {
			return nil, errors.New(fmt.Sprintf("Do not meet the rules of the sql statement %s", string(buffer[1:])))
		}
	}

	return buffer, nil
}

// ResponseBuffer when is local and will call returnCallback function
func setResponseBuffer(iserr bool, buffer []byte, filterCallback parser.Callback) (b []byte) {
	if len(buffer) > 0 && string(buffer[0]) == "Q" {
		if !filterCallback(buffer) {
			return nil
		}
	}

	return buffer
}
