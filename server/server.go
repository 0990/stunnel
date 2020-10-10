package server

import (
	"crypto/cipher"
	"fmt"
	"github.com/0990/stunnel/crypto"
	"github.com/0990/stunnel/tun"
	"github.com/sirupsen/logrus"
	"io"
	"net"
)

func relayToTarget(src tun.Stream, targetAddr string, aead cipher.AEAD) {
	target, err := net.Dial("tcp", targetAddr)
	if err != nil {
		target.Close()
		return
	}

	defer src.Close()
	defer target.Close()
	logrus.Println("stream opened", "in:", fmt.Sprint(src.RemoteAddr(), "(", src.ID(), ")"), "out:", src.RemoteAddr())
	defer logrus.Println("stream closed", "in:", fmt.Sprint(src.RemoteAddr(), "(", src.ID(), ")"), "out:", src.RemoteAddr())

	p3 := crypto.NewConn(src, aead)
	go func() {
		io.Copy(p3, target)
	}()

	_, err = io.Copy(target, p3)
}
