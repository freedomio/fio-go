package main

import (
	"flag"
	"log"

	"github.com/freedomio/fio-go"
)

var listenAddr = flag.String("l", "127.0.0.1:8087", "socks5 listen address")
var httpLisAddr = flag.String("lh", "127.0.0.1:8887", "http listen address")
var remoteAddr = flag.String("r", "8.8.8.8:8087", "remote address")

type combine struct {
	httpProxy  *fio.Http
	quicClient *fio.QUICClient
}

func newCombine(listenAddr, httpLisAddr, remoteAddr string) (*combine, error) {
	quicClient, err := fio.NewQUICClient(listenAddr, remoteAddr, nil, nil)
	if err != nil {
		return nil, err
	}
	httpProxy, err := fio.NewHttp(httpLisAddr, quicClient)
	if err != nil {
		return nil, err
	}
	return &combine{
		httpProxy:  httpProxy,
		quicClient: quicClient,
	}, nil
}

func (c *combine) run() {
	go c.httpProxy.Run()
	c.quicClient.Run()
}

func main() {
	flag.Parse()
	client, err := newCombine(*listenAddr, *httpLisAddr, *remoteAddr)
	if err != nil {
		panic(err)
	}
	log.Println("start client")
	client.run()
}
