package fio

import (
	"io"

	"golang.org/x/sync/errgroup"
)

type proxy struct{}

func (p proxy) transfer(src, dst io.ReadWriter, buf1, buf2 []byte) error {
	var eg errgroup.Group
	eg.Go(func() error {
		_, err := io.CopyBuffer(src, dst, buf1)
		return err
	})

	eg.Go(func() error {
		_, err := io.CopyBuffer(dst, src, buf2)
		return err
	})
	return eg.Wait()
}
