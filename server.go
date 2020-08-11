package stunnel

import (
	"crypto/cipher"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
	"io"
	"net"
)

type Server interface {
	Run() error
}

func NewServer(cfg SConfig) Server {
	return newServer(cfg)
}

type server struct {
	listener net.Listener
	cfg      SConfig

	aead cipher.AEAD
}

func newServer(cfg SConfig) *server {
	aead, err := createAesGcmAEAD(kdf(cfg.AuthKey, 32))
	if err != nil {
		panic(err)
	}
	p := &server{
		cfg:  cfg,
		aead: aead,
	}
	return p
}

func (p *server) Run() error {
	lis, err := kcp.ListenWithOptions(p.cfg.Listen, nil, 30, 15)
	if err != nil {
		return err
	}

	go p.serve(lis)
	//go runUDPRelayServer(p.listenAddr, time.Duration(p.cfg.UDPTimout)*time.Second)
	return nil
}

func (p *server) serve(lis *kcp.Listener) {
	for {
		conn, err := lis.AcceptKCP()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			return
		}
		conn.SetStreamMode(true)
		conn.SetWriteDelay(false)
		conn.SetNoDelay(0, 20, 2, 1)
		conn.SetMtu(1200)
		conn.SetWindowSize(2048, 256)
		conn.SetACKNoDelay(false)

		go p.handleMux(conn)
	}
}

func (p *server) handleMux(conn net.Conn) {
	smuxConfig := smux.DefaultConfig()
	//smuxConfig.MaxReceiveBuffer = 4194304 * 2
	mux, err := smux.Server(conn, smuxConfig)
	if err != nil {
		return
	}

	defer mux.Close()

	for {
		stream, err := mux.AcceptStream()
		if err != nil {
			return
		}

		go func(p1 *smux.Stream) {
			p2, err := net.Dial("tcp", p.cfg.Target)
			if err != nil {
				p1.Close()
				return
			}
			p.handleClient(p1, p2)
		}(stream)
	}
}

func (p *server) handleClient(src *smux.Stream, dst net.Conn) {
	defer src.Close()
	defer dst.Close()
	logrus.Println("stream opened", "in:", fmt.Sprint(src.RemoteAddr(), "(", src.ID(), ")"), "out:", src.RemoteAddr())
	defer logrus.Println("stream closed", "in:", fmt.Sprint(src.RemoteAddr(), "(", src.ID(), ")"), "out:", src.RemoteAddr())

	p3 := newCipherConn(src, p.aead)
	go func() {
		_, err := io.Copy(p3, dst)
		fmt.Println(err)
	}()

	_, err := io.Copy(dst, p3)
	fmt.Println(err)
}
