package process

import (
	"log"
	"net"
	"simplevpn/crypt"
	"simplevpn/tun"
)

type Client struct {
	tunPacket  []byte
	udpPacket  []byte
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
	cryptor    crypt.Cryptor
}

func NewClient(local_addr string, remote_addr string, cryptor crypt.Cryptor) (*Client, error) {
	loacladdr, err := net.ResolveUDPAddr("udp", local_addr)
	if err != nil {
		return nil, err
	}
	remoteaddr, err := net.ResolveUDPAddr("udp", remote_addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", loacladdr)
	if err != nil {
		return nil, err
	}

	return &Client{
		tunPacket:  make([]byte, MTU),
		udpPacket:  make([]byte, MTU),
		conn:       conn,
		remoteAddr: remoteaddr,
		cryptor:    cryptor,
	}, nil
}

func (c *Client) Run() {
	go c.readTunToRemote()
	go c.readRemoteToTun()
}

func (c *Client) Shutdown() {
	c.conn.Close()
}

func (c *Client) readTunToRemote() {
	for {
		// 从tun读取数据
		n, err := tun.IFace.Read(c.tunPacket)
		if err != nil {
			log.Println("tun read error:", err)
			continue
		}
		if n == 0 {
			continue
		}
		// 截取有效数据
		data := c.tunPacket[:n]

		// 加密
		data, err = c.cryptor.Encrypt(data)
		if err != nil {
			log.Println("encrypt data error:", err)
			continue
		}

		// 发送给远程
		_, err = c.conn.WriteToUDP(data, c.remoteAddr)
		if err != nil {
			log.Println("send to remote error:", err)
			continue
		}

	}
}

func (c *Client) readRemoteToTun() {
	for {
		n, err := c.conn.Read(c.udpPacket)
		if err != nil {
			log.Println("read from remote error:", err)
			continue
		}
		if n == 0 {
			continue
		}
		// 截取有效数据
		data := c.udpPacket[:n]

		// 解密
		data, err = c.cryptor.Decrypt(data)
		if err != nil {
			log.Println("decrypt data error:", err)
			continue
		}
		// 写入tun
		_, err = tun.IFace.Write(data)
		if err != nil {
			log.Println("write to tun error:", err)
			continue
		}
	}
}
