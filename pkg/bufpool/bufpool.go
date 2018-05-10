package bufpool

import (
	"sync"
)

var bufPool sync.Pool
var bufCap = 2 * 1024

// Acquire the buffer from pool.
func Acquire() *[]byte {
	return bufPool.Get().(*[]byte)
}

// Giveback the buffer to pool.
func Giveback(buf *[]byte) {
	if cap(*buf) != bufCap {
		panic("Giveback called with packet of wrong size!")
	}
	bufPool.Put(buf)
}

func init() {
	bufPool.New = func() interface{} {
		b := make([]byte, bufCap)
		return &b
	}
}
