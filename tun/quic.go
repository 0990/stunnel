package tun

import (
	"crypto/tls"
	"github.com/0990/stunnel/client"
	"github.com/lucas-clemente/quic-go"
	"net"
)

type QUICTun struct {
	quic.Session
}

func (p *QUICTun) OpenStream() (Stream, error) {
	stream, err := p.Session.OpenStream()
	return &QUICStream{Stream: stream}, err
}

func (p *QUICTun) IsClosed() bool {
	return false
}

type QUICStream struct {
	quic.Stream
}

func (p *QUICStream) ID() int64 {
	return int64(p.Stream.StreamID())
}

func (p *QUICStream) RemoteAddr() net.Addr {
	return nil
}

func NewQUICTun(config client.QUICConfig) (Tun, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-stunnel"},
	}
	session, err := quic.DialAddr(config.Remote, tlsConf, nil)
	if err != nil {
		return nil, err
	}
	return &QUICTun{session: session}, nil
}
