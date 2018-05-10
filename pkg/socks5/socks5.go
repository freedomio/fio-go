package socks5

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

func format(format string) string {
	return fmt.Sprintf("socks5: %s", format)
}

// Socks5 valid socks5 protocol in server side.
func Socks5(rw io.ReadWriter, buf []byte) (string, error) {
	_, err := io.ReadAtLeast(rw, buf, 1)
	if err != nil {
		return "", errors.Wrap(err, format("read version error"))
	}
	version := buf[0]
	if version != socks5Version {
		return "", errors.Errorf(format("unsupport version: %d"), version)
	}
	_, err = rw.Write([]byte{socks5Version, authNone})
	if err != nil {
		return "", errors.Wrap(err, format("write method error"))
	}
	// read request infomation.
	_, err = io.ReadAtLeast(rw, buf, 5)
	if err != nil {
		return "", errors.Wrap(err, format("read request infomation error"))
	}
	if buf[1] != cmdTCPConnect {
		return "", errors.Errorf(format("unsupport cmd: %d"), buf[1])
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
		return "", errors.Errorf(format("unsupport address type: %d"), atype)
	}
	port = binary.BigEndian.Uint16(buf[hostEnd : hostEnd+2])
	_, err = rw.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
	if err != nil {
		return "", errors.Wrap(err, format("write response error"))
	}
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}
