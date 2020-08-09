package stunnel

import (
	"crypto/cipher"
	"fmt"
	"github.com/sirupsen/logrus"
	kcp "github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
	"io"
	"log"
	"net"
	"time"
)

type client struct {
	cfg  CConfig
	aead cipher.AEAD
}

func NewClient(cfg CConfig) *client {

	aead, err := createAesGcmAEAD(kdf(cfg.AuthKey, 32))
	if err != nil {
		panic(err)
	}

	return &client{
		cfg:  cfg,
		aead: aead,
	}
}

func (p *client) Run() error {
	l, err := net.Listen("tcp", p.cfg.LocalAddr)
	if err != nil {
		return err
	}
	go p.serve(l)
	return nil
}

func (p *client) serve(listener net.Listener) {

	mux := struct {
		session *smux.Session
		ttl     time.Time
	}{}

	//mux.session = p.waitConn()
	//mux.ttl = time.Now().Add(time.Duration(config.AutoExpire) * time.Second)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			return
		}
		if mux.session == nil || mux.session.IsClosed() || (time.Now().After(mux.ttl)) {
			mux.session = p.waitConn()
			mux.ttl = time.Now().Add(time.Hour)
		}
		go p.handleClient(mux.session, conn)
	}
}

func (p *client) waitConn() *smux.Session {
	for {
		if session, err := p.createConn(); err == nil {
			return session
		} else {
			logrus.Println("re-connecting:", err)
			time.Sleep(time.Second)
		}
	}
}

func (p *client) createConn() (*smux.Session, error) {
	kcpconn, err := kcp.DialWithOptions(p.cfg.RemoteAddr, nil, 10, 3)
	if err != nil {
		return nil, err
	}
	kcpconn.SetStreamMode(true)
	kcpconn.SetWriteDelay(false)
	kcpconn.SetMtu(1300)
	kcpconn.SetACKNoDelay(true)
	kcpconn.SetNoDelay(1, 10, 2, 1)
	kcpconn.SetWindowSize(512, 2048)

	if err := kcpconn.SetDSCP(46); err != nil {
		log.Println("SetDSCP:", err)
	}
	if err := kcpconn.SetReadBuffer(16777217); err != nil {
		log.Println("SetReadBuffer:", err)
	}
	if err := kcpconn.SetWriteBuffer(16777217); err != nil {
		log.Println("SetWriteBuffer:", err)
	}

	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = 4194304 * 2

	if err := smux.VerifyConfig(smuxConfig); err != nil {
		log.Fatalf("%+v", err)
	}

	session, err := smux.Client(kcpconn, smuxConfig)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (p *client) handleClient(session *smux.Session, p1 net.Conn) {
	defer p1.Close()
	p3, err := session.OpenStream()
	if err != nil {
		logrus.WithError(err).Error("openStream")
		return
	}
	defer p3.Close()

	p2 := newCipherConn(p3, p.aead)

	logrus.Info("stream opened", "in:", p1.RemoteAddr(), "out:", fmt.Sprint(p3.RemoteAddr(), "(", p3.ID(), ")"))
	defer logrus.Info("stream closed", "in:", p1.RemoteAddr(), "out:", fmt.Sprint(p3.RemoteAddr(), "(", p3.ID(), ")"))

	streamCopy := func(dst io.Writer, src io.Reader) {
		if _, err := io.Copy(dst, src); err != nil {
			fmt.Println(err)
			//if err == smux.ErrInvalidProtocol {
			//	logrus.Println("smux", err, "in:", p1.RemoteAddr(), "out:", fmt.Sprint(p2.RemoteAddr(), "(", p2.ID(), ")"))
			//}
		}
		//p1.Close()
		//p2.Close()
	}

	go streamCopy(p1, p2)
	streamCopy(p2, p1)
}
