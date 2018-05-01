package fio

import (
	"sync"
)

var bufferPool sync.Pool
var bufferCap = 2 * 1024

func getPacketBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

func putPacketBuffer(buf *[]byte) {
	if cap(*buf) != bufferCap {
		panic("putPacketBuffer called with packet of wrong size!")
	}
	bufferPool.Put(buf)
}

func init() {
	bufferPool.New = func() interface{} {
		b := make([]byte, bufferCap)
		return &b
	}
}
