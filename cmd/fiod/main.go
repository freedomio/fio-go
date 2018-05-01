package main

import (
	"flag"

	"github.com/freedomio/fio-go"
)

var listenAddr = flag.String("l", "127.0.0.1:8087", "listen address")

func main() {
	flag.Parse()
	server, err := fio.NewServer(*listenAddr)
	if err != nil {
		panic(err)
	}
	server.Run()
}
