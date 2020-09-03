package client

import (
	"context"
	"crypto/tls"
	"github.com/0990/stunnel/tun"
	"github.com/lucas-clemente/quic-go"
)

type QUICTun struct {
	quic.Session
}

func (p *QUICTun) OpenStream() (tun.Stream, error) {
	stream, err := p.Session.OpenStream()
	return &tun.QUICStream{Stream: stream}, err
}

func (p *QUICTun) IsClosed() bool {
	return p.Context().Err() == context.Canceled
}

func newQUICTun(config QUICConfig) (tun.Tun, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-stunnel"},
	}
	session, err := quic.DialAddr(config.Remote, tlsConf, nil)
	if err != nil {
		return nil, err
	}
	return &QUICTun{Session: session}, nil
}
