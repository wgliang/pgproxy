// Copyright 2017 wgliang. All rights reserved.
// Use of this source code is governed by Apache
// license that can be found in the LICENSE file.

// Package cli provides virtual command-line access
// in pgproxy include start,cli and stop action.
package cli

import (
	"os"

	"github.com/golang/glog"
	"github.com/wgliang/pgproxy/filter"
	"github.com/wgliang/pgproxy/proxy"
)

const (
	defaultProxy  = "127.0.0.1:9090"
	defaultRemote = "127.0.0.1:5432"
)

func Main() {
	args := os.Args
	if len(args) > 2 {
		glog.Fatalln("Too many parameters:", args)
		return
	}

	if args[1] == "start" {
		glog.Infoln("Starting pgproxy...")
		proxy.Start(defaultProxy, defaultRemote, filter.GetQueryModificada)
		glog.Infoln("Started pgproxy successfully.")
	} else if args[1] == "cli" {
		// cli
	} else if args[1] == "stop" {
		// stop pgproxy
	} else {
		// print info
	}
}
