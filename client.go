package fio

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/lucas-clemente/quic-go"
)

type Client struct {
	session            quic.Session
	socks5LisAddr      string
	socks5Lis, httpLis net.Listener
}

func NewClient(socks5LisAddr, httpLisAddr, remoteAddr string, tlsCfg *tls.Config) (*Client, error) {
	session, err := quic.DialAddr(remoteAddr, tlsCfg, &quic.Config{
		MaxIncomingStreams:                    65535,
		IdleTimeout:                           365 * 24 * time.Hour,
		MaxReceiveStreamFlowControlWindow:     100 * (1 << 20),
		MaxReceiveConnectionFlowControlWindow: 1000 * (1 << 20),
	})
	if err != nil {
		return nil, err
	}
	socks5Lis, err := net.Listen("tcp", socks5LisAddr)
	if err != nil {
		return nil, err
	}
	httpLis, err := net.Listen("tcp", httpLisAddr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		session:       session,
		socks5LisAddr: socks5LisAddr,
		socks5Lis:     socks5Lis,
		httpLis:       httpLis,
	}
	return c, nil
}

func (c *Client) Run() {
	go c.runSocks5()
	if err := c.runHttp(); err != nil {
		log.Println(err)
	}
}

func (c *Client) runSocks5() {
	for {
		conn, err := c.socks5Lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go c.handleConn(conn)
	}
}
func (c *Client) runHttp() error {
	proxy, err := c.httpBridge()
	if err != nil {
		return err
	}
	return http.Serve(c.httpLis, proxy)
}

func (c *Client) handleConn(conn net.Conn) {
	defer conn.Close()
	buf1 := getPacketBuffer()
	defer putPacketBuffer(buf1)
	// err := defaultSocks5.handshake(conn, *buf1)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// _, err = c.socks5.getAddr(conn, *buf1)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// err = c.socks5.ok(conn)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	buf2 := getPacketBuffer()
	defer putPacketBuffer(buf2)
	stream, err := c.session.OpenStreamSync()
	if err != nil {
		log.Println(err)
		return
	}
	defer stream.Close()
	err = defaultProxy.transfer(conn, stream, *buf1, *buf2)
	if err != nil {
		log.Println(err)
		return
	}
}
