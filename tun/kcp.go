package tun

import (
	"github.com/0990/stunnel/client"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
	"log"
)

type KCPTun struct {
	session *smux.Session
}

func (p *KCPTun) OpenStream() (Stream, error) {
	steam, err := p.session.OpenStream()
	return &KCPStream{Stream: steam}, err
}

func (p *KCPTun) IsClosed() bool {
	return p.session.IsClosed()
}

type KCPStream struct {
	*smux.Stream
}

func (p *KCPStream) ID() int64 {
	return int64(p.Stream.ID())
}

func NewKCPTun(config client.KCPConfig) (Tun, error) {
	kcpConn, err := kcp.DialWithOptions(config.Remote, nil, config.DataShard, config.ParityShard)
	if err != nil {
		return nil, err
	}
	kcpConn.SetStreamMode(true)
	kcpConn.SetWriteDelay(false)
	kcpConn.SetMtu(config.MTU)
	kcpConn.SetACKNoDelay(config.AckNodelay)
	kcpConn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
	kcpConn.SetWindowSize(config.SndWnd, config.Resend)

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
