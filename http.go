package fio

import (
	"log"
	"net"
	"net/http"

	"github.com/elazarl/goproxy"
	xproxy "golang.org/x/net/proxy"
)

type quicForward struct {
	session *quicSession
}

func (f *quicForward) Dial(network, addr string) (c net.Conn, err error) {
	return f.session.OpenConnSync()
}

type Http struct {
	server *http.Server
	client *QUICClient
}

func NewHttp(listenAddr string, client *QUICClient) (*Http, error) {
	socks5Dailer, err := xproxy.SOCKS5("tcp", "0.0.0.0:0", nil, &quicForward{session: client.session})
	if err != nil {
		return nil, err
	}
	httpProxy := goproxy.NewProxyHttpServer()
	httpProxy.ConnectDial = socks5Dailer.Dial
	return &Http{
		server: &http.Server{Addr: listenAddr, Handler: httpProxy},
		client: client,
	}, nil
}

func (h *Http) Run() {
	if err := h.server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
