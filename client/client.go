package client

import (
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
)

type Client struct {
	tunClients   []*tunClient
	rawUDPClient *rawUDPClient
}

func New(cfg TunnelConfig) *Client {
	p := &Client{}

	aead, err := util.CreateAesGcmAead(util.StringToAesKey(cfg.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}

	if cfg.KCP.Listen != "" {
		c, err := NewTunClient("kcp", cfg, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		p.addTunClient(c)
	}

	if cfg.QUIC.Listen != "" {
		c, err := NewTunClient("quic", cfg, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		p.addTunClient(c)
	}

	if cfg.TCP.Listen != "" {
		c, err := NewTunClient("tcp", cfg, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		p.addTunClient(c)
	}

	if cfg.RawUDP.Listen != "" {
		p.rawUDPClient = NewRawUDPClient(cfg.RawUDP, aead)
	}

	return p
}

func (p *Client) addTunClient(c *tunClient) {
	p.tunClients = append(p.tunClients, c)
}

func (p *Client) Run() error {
	for _, v := range p.tunClients {
		err := v.Run()
		if err != nil {
			return err
		}
	}

	if p.rawUDPClient != nil {
		err := p.rawUDPClient.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Client) Close() {
	for _, v := range p.tunClients {
		v.Close()
	}

	if p.rawUDPClient != nil {
		p.rawUDPClient.Close()
	}
}
