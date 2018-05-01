package fio

import (
	"log"
	"net"

	"github.com/lucas-clemente/quic-go"
)

type Client struct {
	session quic.Session
	socks5  socks5
	proxy   proxy
	lis     net.Listener
}

func NewClient(listenAddr, remoteAddr string) (*Client, error) {
	session, err := quic.DialAddr(remoteAddr, nil, nil)
	if err != nil {
		return nil, err
	}
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &Client{
		session: session,
		lis:     lis,
	}, nil
}

func (c *Client) Run() {
	for {
		conn, err := c.lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go c.handleConn(conn)
	}
}

func (c *Client) handleConn(conn net.Conn) {
	buf1 := getPacketBuffer()
	defer putPacketBuffer(buf1)
	err := c.socks5.handshake(conn, *buf1)
	if err != nil {
		log.Println(err)
		return
	}
	err = c.socks5.ok(conn)
	if err != nil {
		log.Println(err)
		return
	}
	buf2 := getPacketBuffer()
	defer putPacketBuffer(buf2)
	stream, err := c.session.OpenStream()
	if err != nil {
		log.Println(err)
		return
	}
	err = c.proxy.transfer(conn, stream, *buf1, *buf2)
	if err != nil {
		log.Println(err)
		return
	}
}
