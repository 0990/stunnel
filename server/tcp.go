package server

import (
	"crypto/cipher"
	"github.com/0990/stunnel/tun"
	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	"net"
)

type tcpServer struct {
	cfg  TCPConfig
	aead cipher.AEAD
}

func newTCPServer(config TCPConfig, aead cipher.AEAD) *tcpServer {
	return &tcpServer{
		cfg:  config,
		aead: aead,
	}
}

func (p *tcpServer) run() error {
	lis, err := net.Listen("tcp", p.cfg.Listen)
	if err != nil {
		return err
	}
	go p.serve(lis)
	return nil
}

func (p *tcpServer) serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
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
