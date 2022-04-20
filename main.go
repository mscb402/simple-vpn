package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"simplevpn/process"
	"simplevpn/tun"
	"syscall"
)

var tunName = flag.String("tun", "cbtun", "tun设备名")
var tunAddr = flag.String("tun_addr", "192.168.9.10/24", "tun绑定的地址")
var remote = flag.String("remote", "", "远程服务器地址")
var listen = flag.String("listen", ":9786", "监听地址")
var mode = flag.String("mode", "server", "模式：server/client")

func main() {
	flag.Parse()
	fmt.Println("程序启动")
	tun.TunName = *tunName
	tun.TunAddr = *tunAddr
	tun.TurnOnTun()

	var p process.Process
	switch *mode {
	case "server":
		server, err := process.NewServer(*listen)
		if err != nil {
			panic(err)
		}
		p = server
	case "client":
		client, err := process.NewClient(*listen, *remote)
		if err != nil {
			panic(err)
		}
		p = client
	}

	p.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	tun.ShutdownTun()
	p.Shutdown()
	fmt.Println("程序退出")
}
