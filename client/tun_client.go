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

type tunClient struct {
	aead       cipher.AEAD
	listenAddr string
	listener   net.Listener
	connNum    int32
	newTun     func() (tun.Tun, error)
}

func NewTunClient(typ string, cfg TunnelConfig, aead cipher.AEAD) (*tunClient, error) {
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

	return &tunClient{
		aead:       aead,
		listenAddr: listen,
		newTun:     newTun,
		connNum:    cfg.ConnNum,
	}, nil
}

func (p *tunClient) Run() error {
	l, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return err
	}

	p.listener = l
	go p.server()
	return nil
}

func (p *tunClient) Close() error {
	return p.listener.Close()
}

func (p *tunClient) waitCreateTun() tun.Tun {
	for {
		logrus.Warn("creating conn....")
		if conn, err := p.newTun(); err == nil {
			logrus.Warn("creating conn ok")
			return conn
		} else {
			logrus.Warn("re-connecting:", err)
			time.Sleep(time.Second)
		}
	}
}

func (p *tunClient) server() {
	numConn := uint16(p.connNum)
	tuns := make([]tun.Tun, numConn)

	for k := range tuns {
		tuns[k] = p.waitCreateTun()
	}
	var tempDelay time.Duration
	rr := uint16(0)
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("server Accept")
			if ne, ok := err.(*net.OpError); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logrus.Errorf("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}

		idx := rr % numConn
		now := time.Now()
		if tuns[idx] == nil || tuns[idx].IsClosed() {
			printElapse("choose tun", now)
			tuns[idx] = p.waitCreateTun()
			printElapse("waitCreateTun", now)
		}

		go p.relayToTun(conn, tuns[idx])
		rr++
	}
}

func printElapse(title string, t time.Time) {
	ms := int32(time.Since(t).Milliseconds())
	logrus.Debugf("%s,elapse:%v", title, ms)
}

func (p *tunClient) relayToTun(src net.Conn, tun tun.Tun) {
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
