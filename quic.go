package fio

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/freedomio/fio-go/pkg/bufpool"
	"github.com/freedomio/fio-go/pkg/socks5"
	"github.com/freedomio/fio-go/pkg/transfer"
	"github.com/lucas-clemente/quic-go"
	"golang.org/x/sync/singleflight"
)

var defaultQuicCfg = &quic.Config{
	IdleTimeout:                           time.Minute * 10,
	MaxIncomingStreams:                    65535,
	MaxReceiveStreamFlowControlWindow:     100 * (1 << 20),
	MaxReceiveConnectionFlowControlWindow: 1000 * (1 << 20),
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

type quicConn struct {
	quic.Stream
	laddr net.Addr
	raddr net.Addr
}

func (c *quicConn) LocalAddr() net.Addr {
	return c.laddr
}

func (c *quicConn) RemoteAddr() net.Addr {
	return c.laddr
}

type quicSession struct {
	quic.Session
}

func (s *quicSession) OpenConnSync() (*quicConn, error) {
	stream, err := s.OpenStreamSync()
	if err != nil {
		return nil, err
	}
	return &quicConn{
		Stream: stream,
		laddr:  s.LocalAddr(),
		raddr:  s.RemoteAddr(),
	}, nil
}

func (s *quicSession) AcceptConn() (*quicConn, error) {
	stream, err := s.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &quicConn{
		Stream: stream,
		laddr:  s.LocalAddr(),
		raddr:  s.RemoteAddr(),
	}, nil
}

type QUICClient struct {
	quicRemoteAddr string
	tlsCfg         *tls.Config
	quicCfg        *quic.Config
	session        *quicSession
	lis            net.Listener
	sf             singleflight.Group
}

func NewQUICClient(listenAddr, quicRemoteAddr string, tlsCfg *tls.Config, quicCfg *quic.Config) (*QUICClient, error) {
	if tlsCfg == nil {
		tlsCfg = &tls.Config{InsecureSkipVerify: true}
	}
	if quicCfg == nil {
		quicCfg = defaultQuicCfg
	}
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	c := &QUICClient{
		quicCfg:        quicCfg,
		tlsCfg:         tlsCfg,
		quicRemoteAddr: quicRemoteAddr,
		lis:            lis,
	}
	c.connect()
	return c, nil
}

func (c *QUICClient) dail() error {
	udpAddr, err := net.ResolveUDPAddr("udp", c.quicRemoteAddr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return err
	}
	session, err := quic.Dial(udpConn, udpAddr, c.quicRemoteAddr, c.tlsCfg, c.quicCfg)
	if err != nil {
		return err
	}
	c.session = &quicSession{Session: session}
	return nil
}

func (c *QUICClient) connect() {
	c.sf.Do("connect", func() (interface{}, error) {
		for {
			err := c.dail()
			if err != nil {
				log.Println("dial quic server error: ", err)
				time.Sleep(3 * time.Second)
				continue
			}
			return nil, nil
		}
	})
}

func (c *QUICClient) Run() {
	for {
		conn, err := c.lis.Accept()
		if err != nil {
			log.Println("tcp accept error: ", err)
			continue
		}
		go c.handleConn(conn)
	}
}

func (c *QUICClient) openConnSync() *quicConn {
	for {
		conn, err := c.session.OpenConnSync()
		if err != nil {
			log.Printf("open conn error: %v\n reconnecting...", err)
			c.connect()
			continue
		}
		return conn
	}
}

func (c *QUICClient) handleConn(in net.Conn) {
	defer in.Close()

	conn := c.openConnSync()
	defer conn.Close()

	buf1 := bufpool.Acquire()
	buf2 := bufpool.Acquire()

	err := transfer.TransferBuffer(conn, in, *buf1, *buf2)
	if err != nil {
		log.Println("transfer error: ", err)
	}
	bufpool.Giveback(buf1)
	bufpool.Giveback(buf2)
}

type QUICServer struct {
	lis quic.Listener
}

func NewQUICServer(listenAddr string, tlsCfg *tls.Config, quicCfg *quic.Config) (*QUICServer, error) {
	if tlsCfg == nil {
		tlsCfg = generateTLSConfig()
	}
	if quicCfg == nil {
		quicCfg = defaultQuicCfg
	}
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	lis, err := quic.Listen(conn, tlsCfg, quicCfg)
	if err != nil {
		return nil, err
	}
	return &QUICServer{
		lis: lis,
	}, nil
}

func (s *QUICServer) Run() {
	for {
		session, err := s.lis.Accept()
		if err != nil {
			log.Println("session accpet error: ", err)
			continue
		}
		log.Println("new session from: ", session.RemoteAddr())
		go s.handleSession(&quicSession{Session: session})
	}
}

func (s *QUICServer) handleSession(session *quicSession) {
	for {
		conn, err := session.AcceptConn()
		if err != nil {
			log.Printf("stream accept error: %v finish session: %s\n", err, session.RemoteAddr())
			return
		}
		go s.handleConn(conn)
	}
}

func (s *QUICServer) handleConn(in io.ReadWriteCloser) {
	defer in.Close()

	buf1 := bufpool.Acquire()
	defer bufpool.Giveback(buf1)
	addr, err := socks5.Socks5(in, *buf1)
	if err != nil {
		log.Println("parse socks5 error: ", err)
		return
	}
	log.Println("new request to: ", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Println("resolve tpc address error: ", err)
		return
	}
	remote, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Println("dial to remote error: ", err)
		return
	}
	defer remote.Close()

	buf2 := bufpool.Acquire()

	err = transfer.TransferBuffer(remote, in, *buf1, *buf2)
	if err != nil {
		log.Println("transfer error: ", err)
	}
	bufpool.Giveback(buf2)
}
