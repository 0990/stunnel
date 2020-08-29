package tun

import (
	"github.com/0990/stunnel/client"
	"github.com/hashicorp/yamux"
	"net"
)

type TCPTun struct {
	session *yamux.Session
}

func (p *TCPTun) OpenStream() (Stream, error) {
	steam, err := p.session.OpenStream()
	return &TCPStream{Stream: steam}, err
}

func (p *TCPTun) IsClosed() bool {
	return p.session.IsClosed()
}

type TCPStream struct {
	*yamux.Stream
}

func (p *TCPStream) ID() int64 {
	return int64(p.StreamID())
}

func NewTCPTun(config client.TCPConfig) (Tun, error) {
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
