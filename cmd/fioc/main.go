package main

import (
	"crypto/tls"
	"flag"

	"github.com/freedomio/fio-go"
)

var listenAddr = flag.String("l", "127.0.0.1:8087", "listen address")
var remoteAddr = flag.String("r", "8.8.8.8:8087", "remote address")

func main() {
	flag.Parse()
	client, err := fio.NewClient(*listenAddr, *remoteAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		panic(err)
	}
	client.Run()
}
