// Copyright 2017 wgliang. All rights reserved.
// Use of this source code is governed by Apache
// license that can be found in the LICENSE file.

// Package cli provides virtual command-line access
// in pgproxy include start,cli and stop action.
package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/wgliang/pgproxy/parser"
	"github.com/wgliang/pgproxy/proxy"
)

var (
	connStr string
	pc      ProxyConfig
)

// pgproxy Main
func Main(config interface{}, pargs interface{}) {
	var proxyconf = flag.String("config", "pgproxy.conf", "configuration file for pgproxy")

	flag.Parse()

	var args []string
	if nil != config {
		pc, connStr = readConfig(config.(string))
		args = pargs.([]string)
	} else {
		pc, connStr = readConfig(*proxyconf)
		args = os.Args
	}

	if len(args) < 2 {
		glog.Errorln("needed one parameters:", args)
		help()
		return
	} else {
		if args[1] == "start" {
			glog.Infoln("Starting pgproxy...")
			info(pc.ServerConfig.ProxyAddr)
			logDir()
			saveCurrentPid()
			proxy.Start(pc.ServerConfig.ProxyAddr, pc.DB["master"].Addr, parser.Filter, parser.Return)
			glog.Infoln("Started pgproxy successfully.")
		} else if args[1] == "cli" {
			Command()
		} else if args[1] == "stop" {
			stop()
		} else {
			help()
		}
	}
}

// print pgproxy help
func help() {
	fmt.Println("	pgproxy is a proxy-server for database postgresql.")
	fmt.Println("	start :start pgproxy server.")
	fmt.Println("	stop :stop pgproxy server.")
	fmt.Println("	version :pgproxy version.")
	fmt.Println("	info :pgproxy info.")
}

// print pgproxy infomation
func info(proxyhost string) {
	fmt.Println(Logo)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "<unknown>"
	}
	pid := strconv.Itoa(os.Getpid())
	starttime := time.Now().Format("2006-01-02 03:04:05 PM")
	fmt.Println("		", VERSION)
	fmt.Println("	Host: " + hostname)
	fmt.Println("	Pid:", string(pid))
	fmt.Println("	Proxy:", proxyhost)
	fmt.Println("	Starttime:", starttime)
	fmt.Println()
}

// set log dir
func logDir() {
	_, err := os.Stat("./log")
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll("./log", 0777)
		if err != nil {
			glog.Fatalln(err)
		} else {
			glog.Infoln("glog and process pid in ./log")
		}
	}
}

// save current pgproxy pid
func saveCurrentPid() {
	// pid file
	filepath := "./log/pid.log"
	fout, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		glog.Errorln(err)
		return
	}
	defer fout.Close()
	// write current pid
	fout.WriteString(strconv.Itoa(os.Getpid()))
}

// get current pgproxy pid
func getCurrentPid() int {
	// pid file
	filepath := "./log/pid.log"
	fin, err := os.OpenFile(filepath, os.O_RDONLY, 0777)
	if err != nil {
		glog.Errorln(err)
		return 0
	}
	defer fin.Close()
	// read current pid
	buf := make([]byte, 1024)

	n, _ := fin.Read(buf)
	if 0 >= n {
		return 0
	} else {
		pid, err := strconv.Atoi(string(buf[0:n]))
		if err != nil {
			glog.Errorln(err)
			return 0
		} else {
			return pid
		}
	}

	return 0
}

// stop pgproxy
func stop() {
	pid := getCurrentPid()
	if pid != 0 {
		err := syscall.Kill(pid, syscall.SIGTERM)
		if err != nil {
			glog.Errorln(err)
		} else {
			glog.Infoln("pgproxy exit successfully!")
		}
	}
	fmt.Printf("pgproxy(%d) Exit,thanks.\n", pid)
}
