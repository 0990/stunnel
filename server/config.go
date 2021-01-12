package server

type Config struct {
	Tunnels  []TunnelConfig `json:"tunnels"`
	LogLevel string         `json:"log_level"`
}

type TunnelConfig struct {
	AuthKey string `json:"authkey"`

	KCP    KCPConfig    `json:"kcp"`
	QUIC   QUICConfig   `json:"quic"`
	TCP    TCPConfig    `json:"tcp"`
	RawUDP RawUDPConfig `json:"rawudp"`
}

type KCPConfig struct {
	Listen       string `json:"listen"`
	Remote       string `json:"remote"`
	MTU          int    `json:"mtu"`
	SndWnd       int    `json:"sndwnd"`
	RcvWnd       int    `json:"rcvwnd"`
	DataShard    int    `json:"datashard"`
	ParityShard  int    `json:"parityshard"`
	DSCP         int    `json:"dscp"`
	AckNodelay   bool   `json:"acknodelay"`
	NoDelay      int    `json:"nodelay"`
	Interval     int    `json:"interval"`
	Resend       int    `json:"resend"`
	NoCongestion int    `json:"nc"`
	SockBuf      int    `json:"sockbuf"`
	StreamBuf    int    `json:"streambuf"`
}

type QUICConfig struct {
	Listen string `json:"listen"`
	Remote string `json:"remote"`
}

type TCPConfig struct {
	Listen string `json:"listen"`
	Remote string `json:"remote"`
}

type RawUDPConfig struct {
	Listen  string `json:"listen"`
	Remote  string `json:"remote"`
	Timeout int    `json:"timeout"`
}
