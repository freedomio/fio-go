package fio

import (
	"log"
	"net"

	"github.com/lucas-clemente/quic-go"
)

type Server struct {
	session quic.Session
	proxy   proxy
	socks5  socks5
}

func NewServer(listenAddr string) (*Server, error) {
	session, err := quic.DialAddr(listenAddr, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Server{
		session: session,
	}, nil
}

func (c *Server) Run() {
	for {
		steam, err := c.session.AcceptStream()
		if err != nil {
			log.Println(err)
			continue
		}
		go c.handleStream(steam)
	}
}

func (c *Server) handleStream(stream quic.Stream) {
	buf1 := getPacketBuffer()
	defer putPacketBuffer(buf1)

	addr, err := c.socks5.getAddr(stream, *buf1)
	if err != nil {
		log.Println(err)
		return
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(err)
		return
	}
	buf2 := getPacketBuffer()
	defer putPacketBuffer(buf2)

	err = c.proxy.transfer(stream, conn, *buf1, *buf2)
	if err != nil {
		log.Println(err)
		return
	}
}
