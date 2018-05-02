package main

import (
	"crypto/tls"
	"flag"

	"github.com/freedomio/fio-go"
)

var socks5LisAddr = flag.String("l", "127.0.0.1:8087", "socks5 listen address")
var httpLisAddr = flag.String("lh", "127.0.0.1:8887", "http listen address")
var remoteAddr = flag.String("r", "8.8.8.8:8087", "remote address")

func main() {
	flag.Parse()
	client, err := fio.NewClient(*socks5LisAddr, *httpLisAddr, *remoteAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		panic(err)
	}
	client.Run()
}
