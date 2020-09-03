package tun

import (
	"github.com/xtaci/smux"
)

type KCPStream struct {
	*smux.Stream
}

func (p *KCPStream) ID() int64 {
	return int64(p.Stream.ID())
}
