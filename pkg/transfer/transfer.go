package transfer

import (
	"io"

	"golang.org/x/sync/errgroup"
)

// TransferBuffer between two io stream with provided buffers.
func TransferBuffer(dst, src io.ReadWriter, buf1, buf2 []byte) error {
	var eg errgroup.Group
	eg.Go(func() error {
		_, err := io.CopyBuffer(dst, src, buf1)
		return err
	})
	eg.Go(func() error {
		_, err := io.CopyBuffer(src, dst, buf2)
		return err
	})
	return eg.Wait()
}
