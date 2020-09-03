package client

import (
	"github.com/0990/stunnel/tun"
	"github.com/hashicorp/yamux"
	"net"
)

type TCPTun struct {
	session *yamux.Session
}

func (p *TCPTun) OpenStream() (tun.Stream, error) {
	steam, err := p.session.OpenStream()
	return &tun.TCPStream{Stream: steam}, err
}

func (p *TCPTun) IsClosed() bool {
	return p.session.IsClosed()
}

func newTCPTun(config TCPConfig) (tun.Tun, error) {
	addr := config.Remote
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}
	return &TCPTun{session: session}, nil
}
