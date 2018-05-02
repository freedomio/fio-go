package fio

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

const (
	socks5Version = 0x05
	authNone      = 0x00
	cmdTCPConnect = 0x01

	atypeIPV4   = 0x01
	atypeDomain = 0x03
	atypeIPV6   = 0x04
)

var defaultSocks5 socks5

type socks5 struct{}

func (s socks5) format(format string) string {
	return fmt.Sprintf("socks5: %s", format)
}

func (s socks5) handshake(rw io.ReadWriter, buf []byte) error {
	_, err := io.ReadAtLeast(rw, buf, 1)
	if err != nil {
		return errors.Wrap(err, s.format("read version error"))
	}
	version := buf[0]
	if version != socks5Version {
		return errors.Errorf(s.format("unsupport version: %d"), version)
	}
	_, err = rw.Write([]byte{socks5Version, authNone})
	return err
}

func (s socks5) getAddr(rw io.ReadWriter, buf []byte) (string, error) {
	_, err := io.ReadAtLeast(rw, buf, 5)
	if err != nil {
		return "", errors.Wrap(err, s.format(""))
	}
	if buf[1] != cmdTCPConnect {
		return "", errors.Errorf(s.format("unsupport cmd: %d"), buf[1])
	}

	atype := buf[3]
	var host string
	var port uint16
	var hostEnd int
	switch atype {
	case atypeIPV4:
		hostEnd = 4 + net.IPv4len
		host = net.IP(buf[4:hostEnd]).String()
	case atypeDomain:
		hostEnd = int(buf[4]) + 5
		host = string(buf[5:hostEnd])
	case atypeIPV6:
		hostEnd = 4 + net.IPv6len
		host = net.IP(buf[4:hostEnd]).String()
	default:
		return "", errors.Errorf(s.format("unsupport address type: %d"), atype)
	}
	port = binary.BigEndian.Uint16(buf[hostEnd : hostEnd+2])
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

func (s socks5) ok(rw io.ReadWriter) error {
	_, err := rw.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
	if err != nil {
		return errors.Wrap(err, s.format(""))
	}
	return nil
}
