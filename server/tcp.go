package server

import (
	"crypto/cipher"
	"github.com/0990/stunnel/tun"
	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type tcpServer struct {
	cfg      TCPConfig
	aead     cipher.AEAD
	listener net.Listener
}

func NewTCPServer(config TCPConfig, aead cipher.AEAD) *tcpServer {
	return &tcpServer{
		cfg:  config,
		aead: aead,
	}
}

func (p *tcpServer) Run() error {
	lis, err := net.Listen("tcp", p.cfg.Listen)
	if err != nil {
		return err
	}
	p.listener = lis
	go p.serve()
	return nil
}

func (p *tcpServer) serve() {
	var tempDelay time.Duration
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			if ne, ok := err.(*net.OpError); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logrus.Errorf("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		go p.handleConn(conn)
	}
}

func (p *tcpServer) handleConn(conn net.Conn) {
	session, err := yamux.Server(conn, nil)
	if err != nil {
		return
	}
	defer session.Close()

	for {
		stream, err := session.AcceptStream()
		if err != nil {
			return
		}

		s := &tun.TCPStream{stream}
		go func(p1 tun.Stream) {
			relayToTarget(p1, p.cfg.Remote, p.aead)
		}(s)
	}
}

func (p *tcpServer) Close() error {
	return p.listener.Close()
}
