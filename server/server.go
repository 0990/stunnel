package server

import (
	"crypto/cipher"
	"fmt"
	"github.com/0990/stunnel/crypto"
	"github.com/0990/stunnel/tun"
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
	"io"
	"net"
)

type server struct {
	kcpServer    *kcpServer
	quicServer   *quicServer
	tcpServer    *tcpServer
	rawUDPServer *rawUDPServer
}

func New(config TunnelConfig) *server {
	p := &server{}

	aead, err := util.CreateAesGcmAead(util.StringToAesKey(config.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}

	if config.KCP.Listen != "" {
		p.kcpServer = NewKCPServer(config.KCP, aead)
	}

	if config.QUIC.Listen != "" {
		p.quicServer = NewQUICServer(config.QUIC, aead)
	}

	if config.TCP.Listen != "" {
		p.tcpServer = NewTCPServer(config.TCP, aead)
	}

	if config.RawUDP.Listen != "" {
		p.rawUDPServer = NewRawUDP(config.RawUDP, aead)
	}

	return p
}

func (p *server) Run() error {
	if p.kcpServer != nil {
		err := p.kcpServer.Run()
		if err != nil {
			return err
		}
	}

	if p.quicServer != nil {
		err := p.quicServer.Run()
		if err != nil {
			return err
		}
	}

	if p.tcpServer != nil {
		err := p.tcpServer.Run()
		if err != nil {
			return err
		}
	}

	if p.rawUDPServer != nil {
		err := p.rawUDPServer.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *server) Close() {
	if p.kcpServer != nil {
		p.kcpServer.Close()
	}

	if p.quicServer != nil {
		p.quicServer.Close()
	}

	if p.tcpServer != nil {
		p.tcpServer.Close()
	}

	if p.rawUDPServer != nil {
		p.rawUDPServer.Close()
	}
}

func relayToTarget(src tun.Stream, targetAddr string, aead cipher.AEAD) {
	defer src.Close()

	target, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	defer target.Close()
	logrus.Println("stream opened", "in:", fmt.Sprint(src.RemoteAddr(), "(", src.ID(), ")"), "out:", src.RemoteAddr())
	defer logrus.Println("stream closed", "in:", fmt.Sprint(src.RemoteAddr(), "(", src.ID(), ")"), "out:", src.RemoteAddr())

	p3 := crypto.NewConn(src, aead)
	go func() {
		io.Copy(p3, target)
	}()

	_, err = io.Copy(target, p3)
}
