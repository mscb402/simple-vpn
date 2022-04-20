package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"simplevpn/crypt"
	"simplevpn/process"
	"simplevpn/tun"
	"syscall"
)

var tunName = flag.String("tun", "utun0", "tun设备名")
var tunAddr = flag.String("tun_addr", "192.168.9.10/24", "tun绑定的地址")
var remote = flag.String("remote", "", "远程服务器地址")
var listen = flag.String("listen", ":9786", "监听地址")
var mode = flag.String("mode", "server", "模式：server/client")
var pwd = flag.String("pwd", "password123", "密码")

func main() {
	flag.Parse()
	fmt.Println("程序启动")
	tun.TunName = *tunName
	tun.TunAddr = *tunAddr
	tun.TurnOnTun()
	defer tun.ShutdownTun()

	// 建立一个xor加密器
	crt := crypt.NewXor(*pwd)

	var p process.Process
	switch *mode {
	case "server":
		server, err := process.NewServer(*listen, crt)
		if err != nil {
			panic(err)
		}
		p = server
	case "client":
		client, err := process.NewClient(*listen, *remote, crt)
		if err != nil {
			panic(err)
		}
		p = client
	}

	p.Run()
	defer p.Shutdown()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	fmt.Println("程序退出")
}
