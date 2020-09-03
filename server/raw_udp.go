package server

import (
	"bytes"
	"crypto/cipher"
	"github.com/0990/stunnel/crypto"
	"net"
	"sync"
	"time"
)

const socketBufSize = 64 * 1024

type RawUDP struct {
	ahead cipher.AEAD
	cfg   RawUDPConfig
}

func NewRawUDP(cfg RawUDPConfig, ahead cipher.AEAD) *RawUDP {
	return &RawUDP{
		cfg:   cfg,
		ahead: ahead,
	}
}
func (p *RawUDP) Run() error {
	remoteAddr, err := net.ResolveUDPAddr("udp", p.cfg.Remote)
	if err != nil {
		return err
	}
	relayer, err := net.ListenPacket("udp", p.cfg.Listen)
	if err != nil {
		return err
	}

	timeout := time.Duration(p.cfg.Timeout) * time.Second
	go runUDPRelayServer(relayer, remoteAddr, p.ahead, timeout)
	return nil
}

//send: client->relayer->sender->remote
//receive: client<-relayer<-sender<-remote
func runUDPRelayServer(relayer net.PacketConn, remoteAddr *net.UDPAddr, aead cipher.AEAD, timeout time.Duration) error {
	defer relayer.Close()
	var senders SenderMap

	buf := make([]byte, socketBufSize)
	data := make([]byte, socketBufSize)
	for {
		n, addr, err := relayer.ReadFrom(data)
		if err != nil {
			continue
		}
		n, err = crypto.NewReader(bytes.NewBuffer(data[0:n]), aead).Read(buf)
		if err != nil {
			continue
		}
		saddr := addr.String()
		sender, exist := senders.Get(saddr)
		if !exist {
			sender, err = net.ListenPacket("udp", "")
			if err != nil {
				continue
			}
			senders.Add(addr.String(), sender)

			go func() {
				relayToClient(sender, relayer, addr, timeout, aead)
				if sender := senders.Del(saddr); sender != nil {
					sender.Close()
				}
			}()
		}

		sender.WriteTo(buf[0:n], remoteAddr)
	}
}

func relayToClient(receiver net.PacketConn, relayer net.PacketConn, clientAddr net.Addr, timeout time.Duration, aead cipher.AEAD) error {
	buf := make([]byte, socketBufSize)
	cryptoWriter := crypto.NewWriter(UDPWriterFunc(func(p []byte) (n int, err error) {
		return relayer.WriteTo(p, clientAddr)
	}), aead)
	for {
		receiver.SetReadDeadline(time.Now().Add(timeout))
		n, _, err := receiver.ReadFrom(buf)
		if err != nil {
			return err
		}
		_, err = cryptoWriter.Write(buf[0:n])
		if err != nil {
			return err
		}
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

type SenderMap struct {
	sync.Map
}

func (p *SenderMap) Add(key string, sender net.PacketConn) {
	p.Map.Store(key, sender)
}

func (p *SenderMap) Del(key string) net.PacketConn {
	if conn, exist := p.Get(key); exist {
		p.Map.Delete(key)
		return conn
	}

	return nil
}

func (p *SenderMap) Get(key string) (net.PacketConn, bool) {
	v, exist := p.Load(key)
	if !exist {
		return nil, false
	}

	return v.(net.PacketConn), true
}
