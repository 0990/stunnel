package server

import (
	"crypto/cipher"
	"github.com/0990/stunnel/tun"
	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	"net"
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
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
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
