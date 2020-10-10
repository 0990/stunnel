package client

import (
	"github.com/0990/stunnel/tun"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
	"log"
)

type KCPTun struct {
	session *smux.Session
}

func (p *KCPTun) OpenStream() (tun.Stream, error) {
	steam, err := p.session.OpenStream()
	return &tun.KCPStream{Stream: steam}, err
}

func (p *KCPTun) IsClosed() bool {
	return p.session.IsClosed()
}

func newKCPTun(config KCPConfig) (tun.Tun, error) {
	kcpConn, err := kcp.DialWithOptions(config.Remote, nil, config.DataShard, config.ParityShard)
	if err != nil {
		return nil, err
	}
	kcpConn.SetStreamMode(true)
	kcpConn.SetWriteDelay(false)
	kcpConn.SetMtu(config.MTU)
	kcpConn.SetACKNoDelay(config.AckNodelay)
	kcpConn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
	kcpConn.SetWindowSize(config.SndWnd, config.RcvWnd)

	if err := kcpConn.SetDSCP(config.DSCP); err != nil {
		log.Println("SetDSCP:", err)
	}
	if err := kcpConn.SetReadBuffer(config.SockBuf); err != nil {
		log.Println("SetReadBuffer:", err)
	}
	if err := kcpConn.SetWriteBuffer(config.SockBuf); err != nil {
		log.Println("SetWriteBuffer:", err)
	}

	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = config.StreamBuf

	if err := smux.VerifyConfig(smuxConfig); err != nil {
		log.Fatalf("%+v", err)
	}

	session, err := smux.Client(kcpConn, smuxConfig)
	if err != nil {
		return nil, err
	}
	return &KCPTun{session: session}, nil
}
