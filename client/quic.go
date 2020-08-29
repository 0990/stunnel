package client

import (
	"crypto/cipher"
	"crypto/tls"
	"fmt"
	"github.com/0990/stunnel/crypto"
	"github.com/0990/stunnel/util"
	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type QUIC struct {
	config QUICConfig
	aead   cipher.AEAD
}

func (p *QUIC) Run() error {
	lAddr := p.config.Listen
	l, err := net.Listen("tcp", lAddr)
	if err != nil {
		return err
	}

	go p.server(l)
	return nil
}

func (p *QUIC) server(l net.Listener) {
	var session quic.Session

	for {
		conn, err := l.Accept()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			return
		}
		if session == nil {
			session = p.waitCreateSession()
		}
		go p.relayToSession(conn, session)
	}

}

func (p *QUIC) waitCreateSession() quic.Session {
	for {
		if session, err := p.createSession(); err == nil {
			return session
		} else {
			logrus.Println("re-connecting:", err)
			time.Sleep(time.Second)
		}
	}
}

func (p *QUIC) createSession() (quic.Session, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-stunnel"},
	}
	session, err := quic.DialAddr(p.config.Remote, tlsConf, nil)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (p *QUIC) relayToSession(conn net.Conn, session quic.Session) {
	defer conn.Close()
	steam, err := session.OpenStream()
	if err != nil {
		logrus.WithError(err).Error("openStream")
		return
	}
	defer steam.Close()

	logrus.Info("stream opened", "in:", conn.RemoteAddr(), "out:", fmt.Sprint(session.RemoteAddr(), "(", steam.StreamID(), ")"))
	defer logrus.Info("stream closed", "in:", conn.RemoteAddr(), "out:", fmt.Sprint(session.RemoteAddr(), "(", steam.StreamID(), ")"))

	cipherSess := crypto.NewConn(steam, p.aead)
	go util.Copy(conn, cipherSess)
	util.Copy(cipherSess, conn)
}
