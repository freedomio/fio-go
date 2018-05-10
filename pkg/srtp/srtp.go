package srtp

import (
	"encoding/binary"
	"math/rand"
	"net"
)

// SRTP fake SRTP header.
type SRTP struct {
	net.PacketConn
	header uint16
	number uint16
}

// WriteTo PacketConn WriteTo implement.
func (s *SRTP) WriteTo(b []byte, addr net.Addr) (int, error) {
	s.number++
	header := make([]byte, 4)
	binary.BigEndian.PutUint16(header[0:2], s.number)
	binary.BigEndian.PutUint16(header[2:4], s.number)
	n, err := s.PacketConn.WriteTo(append(header, b...), addr)
	return n - 4, err
}

// ReadFrom PacketConn ReadFrom implement.
func (s *SRTP) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := s.PacketConn.ReadFrom(b)
	b = b[4:]
	return n - 4, addr, err
}

// New returns a new SRTP instance based on the given config.
func New(conn net.PacketConn) *SRTP {
	return &SRTP{
		PacketConn: conn,
		header:     0xB5E8,
		number:     uint16(rand.Intn(65536)),
	}
}
