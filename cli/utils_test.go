package cli

import (
	"testing"
)

func Test_readConfig(t *testing.T) {
	readConfig("../pgproxy.conf")
}
