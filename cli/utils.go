package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	ProxyAddr  string `json:"proxyAddr"`
	RemoteAddr string `json:"remoteAddr"`
	PgUser     string `json:"pgUser"`
	PgPassword string `json:"pgPassword"`
	PgDb       string `json:"pgDb"`
}

// readConfig file
func readConfig(file string) (proxy, remote, connstr string) {
	f, err := os.Open(file)
	if err != nil {
		glog.Errorln(err)
		return
	}

	var p ProxyConfig
	if err = json.NewDecoder(f).Decode(&p); err != nil {
		glog.Errorln(err)
		return
	}

	sepindex := strings.Index(p.ProxyAddr, ":")

	return p.ProxyAddr, p.RemoteAddr, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=skylar sslmode=disable",
		p.ProxyAddr[0:sepindex], p.ProxyAddr[(sepindex+1):], p.PgUser, p.PgPassword)
}
