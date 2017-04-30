// Copyright 2017 wgliang. All rights reserved.
// Use of this source code is governed by Apache
// license that can be found in the LICENSE file.

// Package cli provides virtual command-line access
// in pgproxy include start,cli and stop action.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/bbangert/toml"
	"github.com/golang/glog"
)

const Logo = `
    ____  ____ _____  _________  _  ____  __
   / __ \/ __ '/ __ \/ ___/ __ \| |/_/ / / /
  / /_/ / /_/ / /_/ / /  / /_/ />  </ /_/ / 
 / .___/\__, / .___/_/   \____/_/|_|\__, /  
/_/    /____/_/                    /____/   
`

const (
	VERSION = "version-0.0.1"
)

// proxy server config struct
type ProxyConfig struct {
	ServerConfig struct {
		ProxyAddr string
	}
	DB map[string]struct {
		Addr     string
		User     string
		Password string
		DbName   string
	} `toml:"DB"`
}

func readConfig(file string) (pc ProxyConfig, connStr string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		glog.Fatalln(err)
	}

	if _, err := toml.DecodeFile(file, &pc); err != nil {
		glog.Fatalln(err)
	}

	sepindex := strings.Index(pc.DB["master"].Addr, ":")

	return pc, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s application_name=pgproxy sslmode=disable",
		pc.DB["master"].Addr[0:sepindex], pc.DB["master"].Addr[(sepindex+1):], pc.DB["master"].User, pc.DB["master"].Password, pc.DB["master"].DbName)
}
