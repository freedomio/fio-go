package fio

import (
	"net"

	"github.com/lucas-clemente/quic-go"

	"github.com/elazarl/goproxy"
	xproxy "golang.org/x/net/proxy"
)

type fakeStream struct {
	session quic.Session
	quic.Stream
}

func (s *fakeStream) LocalAddr() net.Addr {
	return s.session.LocalAddr()
}

func (s *fakeStream) RemoteAddr() net.Addr {
	return s.session.RemoteAddr()
}

type quicForward struct {
	session quic.Session
}

func (f *quicForward) Dial(network, addr string) (c net.Conn, err error) {
	stream, err := f.session.OpenStreamSync()
	if err != nil {
		return nil, err
	}
	return &fakeStream{
		session: f.session,
		Stream:  stream,
	}, nil
}

func (c *Client) httpBridge() (*goproxy.ProxyHttpServer, error) {
	socks5Dailer, err := xproxy.SOCKS5("tcp", c.socks5LisAddr, nil, &quicForward{session: c.session})
	if err != nil {
		return nil, err
	}
	httpProxy := goproxy.NewProxyHttpServer()
	httpProxy.ConnectDial = socks5Dailer.Dial
	return httpProxy, nil
}
