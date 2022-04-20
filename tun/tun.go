package tun

import (
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

var TunName = "cbTun"
var TunAddr = ""

type callbackFunc func()

var (
	IFace                        *water.Interface
	closeTurnFunc, closeLinkFunc callbackFunc
)

func TurnOnTun() {
	// 启动tun接口
	closeTurnFunc = initTun()
	// 启动link
	closeLinkFunc = initLink()
}

func ShutdownTun() {
	closeTurnFunc()
	closeLinkFunc()
}

func initTun() callbackFunc {
	// 设置网卡类型、名称
	cfg := water.Config{DeviceType: water.TUN}
	cfg.Name = TunName

	// 监听网卡
	iface, err := water.New(cfg)
	if err != nil {
		panic(err)
	}
	IFace = iface

	return func() {
		err := iface.Close()
		if err != nil {
			panic(err)
		}
	}
}

func initLink() callbackFunc {
	// 设置IP
	link, err := netlink.LinkByName(TunName)
	if err != nil {
		panic(err)
	}

	// 解析地址
	addr, err := netlink.ParseAddr(TunAddr)

	// tun添加地址
	err = netlink.AddrAdd(link, addr)
	if err != nil {
		panic(err)
	}

	// 设置MTU
	err = netlink.LinkSetMTU(link, 1300)
	if err != nil {
		panic(err)
	}

	// 启动
	err = netlink.LinkSetUp(link)
	if err != nil {
		panic(err)
	}

	return func() {
		// 关闭link的回调函数
		err := netlink.LinkSetDown(link)
		if err != nil {
			panic(err)
		}
	}
}
