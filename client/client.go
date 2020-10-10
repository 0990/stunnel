package client

import (
	"crypto/cipher"
	"fmt"
	"github.com/0990/stunnel/crypto"
	"github.com/0990/stunnel/tun"
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type client struct {
	aead    cipher.AEAD
	listen  string
	connNum int32
	newTun  func() (tun.Tun, error)
}

func New(typ string, cfg Config, aead cipher.AEAD) (*client, error) {
	var newTun func() (tun.Tun, error)
	var listen string
	switch typ {
	case "kcp":
		newTun = func() (tun.Tun, error) {
			return newKCPTun(cfg.KCP)
		}
		listen = cfg.KCP.Listen
	case "quic":
		newTun = func() (tun.Tun, error) {
			return newQUICTun(cfg.QUIC)
		}
		listen = cfg.QUIC.Listen
	case "tcp":
		newTun = func() (tun.Tun, error) {
			return newTCPTun(cfg.TCP)
		}
		listen = cfg.TCP.Listen
	default:
		return nil, fmt.Errorf("not support typ:%s", typ)
	}

	return &client{
		aead:    aead,
		listen:  listen,
		newTun:  newTun,
		connNum: cfg.ConnNum,
	}, nil
}

func (p *client) Run() error {
	l, err := net.Listen("tcp", p.listen)
	if err != nil {
		return err
	}

	go p.server(l)
	return nil
}

func (p *client) waitCreateTun() tun.Tun {
	for {
		logrus.Info("creating conn....")
		if conn, err := p.newTun(); err == nil {
			logrus.Info("creating conn ok")
			return conn
		} else {
			logrus.Println("re-connecting:", err)
			time.Sleep(time.Second)
		}
	}
}

func (p *client) server(l net.Listener) {
	numConn := uint16(p.connNum)
	tuns := make([]tun.Tun, numConn)

	for k := range tuns {
		tuns[k] = p.waitCreateTun()
	}

	rr := uint16(0)
	for {
		conn, err := l.Accept()
		if err != nil {
			logrus.WithError(err).Error("server Accept")
			return
		}

		idx := rr % numConn
		if tuns[idx] == nil || tuns[idx].IsClosed() {
			tuns[idx] = p.waitCreateTun()
		}
		go p.relayToTun(conn, tuns[idx])
		rr++
	}
}

func (p *client) relayToTun(src net.Conn, tun tun.Tun) {
	defer src.Close()
	stream, err := tun.OpenStream()
	if err != nil {
		logrus.WithError(err).Error("openStream")
		return
	}
	defer stream.Close()

	logrus.Debug("stream opened", "in:", src.RemoteAddr(), "out:", fmt.Sprint(stream.RemoteAddr(), "(", stream.ID(), ")"))
	defer logrus.Debug("stream closed", "in:", src.RemoteAddr(), "out:", fmt.Sprint(stream.RemoteAddr(), "(", stream.ID(), ")"))

	cipherSess := crypto.NewConn(stream, p.aead)

	go util.Copy(src, cipherSess)
	util.Copy(cipherSess, src)
}
