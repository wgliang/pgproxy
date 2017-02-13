package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wgliang/pgproxy/cli"
)

func main() {
	// call proxy
	cli.Main("../pgproxy.json", []string{"pgproxy", "start"})

	// 捕获ctrl-c,平滑退出
	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-chExit:
		fmt.Println("Example EXITING...Bye.")
	}
}
