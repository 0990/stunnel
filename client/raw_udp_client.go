package client

import (
	"bytes"
	"crypto/cipher"
	"github.com/0990/stunnel/crypto"
	"io"
	"net"
	"sync"
	"time"
)

const socketBufSize = 64 * 1024

type rawUDPClient struct {
	ahead   cipher.AEAD
	cfg     RawUDPConfig
	relayer net.PacketConn
}

func NewRawUDPClient(cfg RawUDPConfig, ahead cipher.AEAD) *rawUDPClient {
	return &rawUDPClient{
		cfg:   cfg,
		ahead: ahead,
	}
}

func (p *rawUDPClient) Run() error {
	remoteAddr, err := net.ResolveUDPAddr("udp", p.cfg.Remote)
	if err != nil {
		return err
	}
	relayer, err := net.ListenPacket("udp", p.cfg.Listen)
	if err != nil {
		return err
	}

	p.relayer = relayer

	timeout := time.Duration(p.cfg.Timeout) * time.Second
	go runUDPRelayServer(relayer, remoteAddr, p.ahead, timeout)
	return nil
}

func (p *rawUDPClient) Close() error {
	return p.relayer.Close()
}

//send: client->relayer->sender->remote
//receive: client<-relayer<-sender<-remote
func runUDPRelayServer(relayer net.PacketConn, remoteAddr *net.UDPAddr, aead cipher.AEAD, timeout time.Duration) error {
	remote, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return err
	}
	defer remote.Close()
	defer relayer.Close()

	var senders SenderMap

	buf := make([]byte, socketBufSize)
	for {
		n, addr, err := relayer.ReadFrom(buf)
		if err != nil {
			continue
		}
		saddr := addr.String()
		sender, exist := senders.Get(saddr)
		if !exist {
			conn, err := net.ListenPacket("udp", "")
			if err != nil {
				continue
			}
			sender = newUDPSender(conn, remoteAddr, aead)
			senders.Add(addr.String(), sender)

			go func() {
				relayToClient(sender.conn, relayer, addr, timeout, aead)
				if sender := senders.Del(saddr); sender != nil {
					sender.conn.Close()
				}
			}()
		}

		sender.Write(buf[0:n])
	}
}

func relayToClient(receiver net.PacketConn, relayer net.PacketConn, clientAddr net.Addr, timeout time.Duration, aead cipher.AEAD) error {
	buf := make([]byte, socketBufSize)
	data := make([]byte, socketBufSize)
	for {
		receiver.SetReadDeadline(time.Now().Add(timeout))

		n, _, err := receiver.ReadFrom(data)
		if err != nil {
			return err
		}

		n, err = crypto.NewReader(bytes.NewBuffer(data[0:n]), aead).Read(buf)
		if err != nil {
			continue
		}

		_, err = relayer.WriteTo(buf[:n], clientAddr)
		if err != nil {
			return err
		}
	}
}

func newUDPSender(conn net.PacketConn, remoteAddr *net.UDPAddr, aead cipher.AEAD) *UDPSender {
	w := crypto.NewWriter(UDPWriterFunc(func(p []byte) (n int, err error) {
		return conn.WriteTo(p, remoteAddr)
	}), aead)

	return &UDPSender{
		conn:       conn,
		remoteAddr: remoteAddr,
		aead:       nil,
		Writer:     w,
	}
}

type UDPWriterFunc func(p []byte) (n int, err error)

func (f UDPWriterFunc) Write(p []byte) (n int, err error) {
	return f(p)
}

type UDPReaderFunc func(p []byte) (n int, err error)

func (f UDPReaderFunc) Read(p []byte) (n int, err error) {
	return f(p)
}

type UDPSender struct {
	conn       net.PacketConn
	remoteAddr *net.UDPAddr
	aead       cipher.AEAD

	io.Writer
	io.Closer
}

func (p *UDPSender) Close() error {
	return p.conn.Close()
}

func (p *UDPSender) SetReadDeadline(t time.Time) {
	p.conn.SetReadDeadline(t)
}

func (p *UDPSender) Write(d []byte) (int, error) {
	return p.Writer.Write(d)
}

type SenderMap struct {
	sync.Map
}

func (p *SenderMap) Add(key string, sender *UDPSender) {
	p.Map.Store(key, sender)
}

func (p *SenderMap) Del(key string) *UDPSender {
	if conn, exist := p.Get(key); exist {
		p.Map.Delete(key)
		return conn
	}

	return nil
}

func (p *SenderMap) Get(key string) (*UDPSender, bool) {
	v, exist := p.Load(key)
	if !exist {
		return nil, false
	}

	return v.(*UDPSender), true
}
