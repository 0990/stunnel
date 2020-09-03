package tun

import (
	"github.com/hashicorp/yamux"
)

type TCPStream struct {
	*yamux.Stream
}

func (p *TCPStream) ID() int64 {
	return int64(p.StreamID())
}
