package process

import (
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"simplevpn/tun"
)

type Server struct {
	tunPacket []byte
	udpPacket []byte
	conn      *net.UDPConn
	peers     map[string]*net.UDPAddr
}

func NewServer(lisenAddr string) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", lisenAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		tunPacket: make([]byte, MTU),
		udpPacket: make([]byte, MTU),
		conn:      conn,
		peers:     make(map[string]*net.UDPAddr),
	}, nil
}
func (s *Server) Run() {
	go s.readTunToRemote()
	go s.readClientToTun()
}
func (s *Server) Shutdown() {
	s.conn.Close()
}

func (s *Server) readTunToRemote() {
	for {
		// 从tun读取数据
		n, err := tun.IFace.Read(s.tunPacket)
		if err != nil {
			log.Println("tun read error:", err)
			continue
		}
		if n == 0 {
			continue
		}
		// 截取有效数据
		data := s.tunPacket[:n]

		// 读取目的地址
		header, err := ipv4.ParseHeader(data)
		if err != nil {
			log.Println("parse header error:", err)
			continue
		}
		dstAddr := header.Dst.String()
		dstClient, ok := s.peers[dstAddr]
		if !ok {
			log.Println("no peer found:", dstAddr)
			continue
		}

		// 加密
		data, err = encryptData(data)
		if err != nil {
			log.Println("encrypt data error:", err)
			continue
		}

		// 发送给远程
		_, err = s.conn.WriteToUDP(data, dstClient)
		if err != nil {
			log.Println("send to remote error:", err)
			continue
		}

	}
}

func (s *Server) readClientToTun() {
	for {
		n, clientAddr, err := s.conn.ReadFromUDP(s.udpPacket)
		if err != nil {
			log.Println("read from remote error:", err)
			continue
		}
		if n == 0 {
			continue
		}
		data := s.udpPacket[:n]
		data, err = decryptData(data)
		if err != nil {
			log.Println("decrypt data error:", err)
			continue
		}

		// 把源IP取出来
		header, err := ipv4.ParseHeader(data)
		if err != nil {
			log.Println("parse ipv4 header error:", err)
			continue
		}
		srcIP := header.Src.String()

		// 如果是第一次收到数据，则把远程的地址保存下来
		// todo 这边的srcIP需要淘汰
		if _, ok := s.peers[srcIP]; !ok {
			s.peers[srcIP] = clientAddr
		}

		// 写入tun
		_, err = tun.IFace.Write(data)
		if err != nil {
			log.Println("write to tun error:", err)
			continue
		}
	}
}
