package stunnel

type CConfig struct {
	LocalAddr  string `json:"localaddr"`
	RemoteAddr string `json:"remoteaddr"`
	AuthKey    string `json:"authkey"`
}

type SConfig struct {
	Listen  string `json:"listen"`
	Target  string `json:"target"`
	AuthKey string `json:"authkey"`
}
