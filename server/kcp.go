package server

import (
	"crypto/cipher"
	"github.com/0990/stunnel/tun"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
	"log"
	"net"
)

type kcpServer struct {
	cfg      KCPConfig
	aead     cipher.AEAD
	listener *kcp.Listener
}

func NewKCPServer(config KCPConfig, aead cipher.AEAD) *kcpServer {
	return &kcpServer{
		cfg:  config,
		aead: aead,
	}
}

func (p *kcpServer) Run() error {
	lis, err := kcp.ListenWithOptions(p.cfg.Listen, nil, p.cfg.DataShard, p.cfg.ParityShard)
	if err != nil {
		return err
	}

	config := p.cfg

	if err := lis.SetDSCP(config.DSCP); err != nil {
		log.Println("SetDSCP:", err)
	}
	if err := lis.SetReadBuffer(config.SockBuf); err != nil {
		log.Println("SetReadBuffer:", err)
	}
	if err := lis.SetWriteBuffer(config.SockBuf); err != nil {
		log.Println("SetWriteBuffer:", err)
	}

	p.listener = lis

	go p.serve()
	return nil
}

func (p *kcpServer) Close() error {
	return p.listener.Close()
}

func (p *kcpServer) serve() {
	config := p.cfg
	for {
		conn, err := p.listener.AcceptKCP()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			return
		}
		conn.SetStreamMode(true)
		conn.SetWriteDelay(false)
		conn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
		conn.SetMtu(config.MTU)
		conn.SetWindowSize(config.SndWnd, config.RcvWnd)
		conn.SetACKNoDelay(config.AckNodelay)

		go p.handleConn(conn)
	}
}

func (p *kcpServer) handleConn(conn net.Conn) {
	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = p.cfg.StreamBuf
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

		s := &tun.KCPStream{Stream: stream}
		go func(s tun.Stream) {
			relayToTarget(s, p.cfg.Remote, p.aead)
		}(s)
	}
}
