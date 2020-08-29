package server

type Config struct {
	AuthKey string `json:"authkey"`

	KCP  KCPConfig  `json:"kcp"`
	QUIC QUICConfig `json:"quic"`
	TCP  TCPConfig  `json:"tcp"`
}

type KCPConfig struct {
	Listen       string `json:"listen"`
	Remote       string `json:"remote"`
	Key          string `json:"key"`
	Crypt        string `json:"crypt"`
	MTU          int    `json:"mtu"`
	SndWnd       int    `json:"sndwnd"`
	RcvWnd       int    `json:"rcvwnd"`
	DataShard    int    `json:"datashard"`
	ParityShard  int    `json:"parityshard"`
	DSCP         int    `json:"dscp"`
	NoComp       bool   `json:"nocomp"`
	AckNodelay   bool   `json:"acknodelay"`
	NoDelay      int    `json:"nodelay"`
	Interval     int    `json:"interval"`
	Resend       int    `json:"resend"`
	NoCongestion int    `json:"nc"`
	SockBuf      int    `json:"sockbuf"`
	StreamBuf    int    `json:"streambuf"`
	KeepAlive    int    `json:"keepalive"`
	TCP          bool   `json:"tcp"`
}

type QUICConfig struct {
	Listen string `json:"listen"`
	Remote string `json:"remote"`
}

type TCPConfig struct {
	Listen string `json:"listen"`
	Remote string `json:"remote"`
}
