package tun

import (
	"io"
	"net"
)

type Tun interface {
	OpenStream() (Stream, error)
	IsClosed() bool
}

type Stream interface {
	ID() int64
	RemoteAddr() net.Addr
	io.ReadWriteCloser
}
