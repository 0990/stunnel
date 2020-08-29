package server

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/0990/stunnel/tun"
	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
	"math/big"
)

type quicServer struct {
	cfg  QUICConfig
	aead cipher.AEAD
}

func newQUICServer(config QUICConfig, aead cipher.AEAD) *quicServer {
	return &quicServer{
		cfg:  config,
		aead: aead,
	}
}

func (p *quicServer) run() error {
	lis, err := quic.ListenAddr(p.cfg.Listen, generateTLSConfig(), nil)
	if err != nil {
		return err
	}
	go p.serve(lis)
	return nil
}

func (p *quicServer) serve(lis quic.Listener) {
	for {

		sess, err := lis.Accept(context.Background())
		if err != nil {
			logrus.WithError(err).Error("quicServer Accept")
			return
		}
		go p.handleSession(sess)
	}
}

func (p *quicServer) handleSession(session quic.Session) {
	defer session.CloseWithError(1, "quic server close session")

	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			return
		}

		s := &tun.QUICStream{Stream: stream}
		go func(s tun.Stream) {
			relayToTarget(s, p.cfg.Remote, p.aead)
		}(s)
	}
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:    "CERTIFICATE",
		Headers: nil,
		Bytes:   certDER,
	})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
