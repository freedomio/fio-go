package fio

import (
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
)

type Server struct {
	lis    quic.Listener
	proxy  proxy
	socks5 socks5
}

func NewServer(listenAddr string, tlsCfg *tls.Config) (*Server, error) {
	lis, err := quic.ListenAddr(listenAddr, tlsCfg, &quic.Config{
		MaxIncomingStreams:                    65535,
		IdleTimeout:                           365 * 24 * time.Hour,
		MaxReceiveStreamFlowControlWindow:     100 * (1 << 20),
		MaxReceiveConnectionFlowControlWindow: 1000 * (1 << 20),
	})
	if err != nil {
		return nil, err
	}
	return &Server{
		lis: lis,
	}, nil
}

func (c *Server) Run() {
	for {
		session, err := c.lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("new session: %s\n", session.RemoteAddr())
		go c.handleSession(session)
	}
}

func (c *Server) handleSession(session quic.Session) {
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			log.Println(err)
			continue
		}
		go c.handleStream(stream)
	}
}

func (c *Server) handleStream(stream quic.Stream) {
	defer stream.Close()
	buf1 := getPacketBuffer()
	defer putPacketBuffer(buf1)

	addr, err := c.socks5.getAddr(stream, *buf1)
	if err != nil {
		log.Println(err)
		return
	}
	err = c.socks5.ok(stream)
	if err != nil {
		log.Println(err)
		return
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	buf2 := getPacketBuffer()
	defer putPacketBuffer(buf2)

	err = c.proxy.transfer(stream, conn, *buf1, *buf2)
	if err != nil {
		log.Println(err)
		return
	}
}
