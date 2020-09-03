package tun

import (
	"github.com/lucas-clemente/quic-go"
	"net"
)

type QUICStream struct {
	quic.Stream
}

func (p *QUICStream) ID() int64 {
	return int64(p.Stream.StreamID())
}

func (p *QUICStream) RemoteAddr() net.Addr {
	return nil
}
