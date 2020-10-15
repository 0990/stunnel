package stunnel

import (
	"encoding/json"
	"fmt"
	"github.com/0990/stunnel/client"
	"github.com/0990/stunnel/server"
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

func Test_CreateConfig(t *testing.T) {
	cCfg := client.Config{
		AuthKey: "abcdefg",
		ConnNum: 10,
		KCP: client.KCPConfig{
			Listen:       "0.0.0.0:1000",
			Remote:       ":1001",
			MTU:          1200,
			SndWnd:       256,
			RcvWnd:       2048,
			DataShard:    30,
			ParityShard:  15,
			DSCP:         46,
			AckNodelay:   false,
			NoDelay:      0,
			Interval:     20,
			Resend:       2,
			NoCongestion: 1,
			SockBuf:      16777217,
			StreamBuf:    4194304,
		},
		QUIC: client.QUICConfig{
			Listen: "0.0.0.0:2000",
			Remote: ":2001",
		},
		TCP: client.TCPConfig{
			Listen: "0.0.0.0:3000",
			Remote: ":3001",
		},
		RawUDP: client.RawUDPConfig{
			Listen:  "0.0.0.0:4000",
			Remote:  ":4001",
			Timeout: 0,
		},
	}

	sCfg := server.Config{
		AuthKey: "abcdefg",
		KCP: server.KCPConfig{
			Listen:       "0.0.0.0:1001",
			Remote:       "",
			MTU:          1200,
			SndWnd:       2048,
			RcvWnd:       256,
			DataShard:    30,
			ParityShard:  15,
			DSCP:         46,
			AckNodelay:   false,
			NoDelay:      0,
			Interval:     20,
			Resend:       2,
			NoCongestion: 1,
			SockBuf:      16777217,
			StreamBuf:    4194304,
		},
		QUIC: server.QUICConfig{
			Listen: "0.0.0.0:2001",
			Remote: "",
		},
		TCP: server.TCPConfig{
			Listen: "0.0.0.0:3001",
			Remote: "",
		},
		RawUDP: server.RawUDPConfig{
			Listen:  "0.0.0.0:4001",
			Remote:  ":4001",
			Timeout: 0,
		},
	}

	c, _ := json.MarshalIndent(cCfg, "", "   ")
	ioutil.WriteFile("stclient.json", c, 0644)

	s, _ := json.MarshalIndent(sCfg, "", "   ")
	ioutil.WriteFile("stserver.json", s, 0644)
}

func Test_KCP(t *testing.T) {
	targetAddr := "127.0.0.1:3000"
	localProxyAddr := "127.0.0.1:1000"

	cCfg := client.Config{
		AuthKey: "abcdefg",
		ConnNum: 10,
		KCP: client.KCPConfig{
			Listen:       localProxyAddr,
			Remote:       "127.0.0.1:1001",
			MTU:          1200,
			SndWnd:       256,
			RcvWnd:       2048,
			DataShard:    30,
			ParityShard:  15,
			DSCP:         46,
			AckNodelay:   true,
			NoDelay:      1,
			Interval:     10,
			Resend:       2,
			NoCongestion: 1,
			SockBuf:      16777217,
			StreamBuf:    4194304,
		},
		QUIC: client.QUICConfig{},
		TCP:  client.TCPConfig{},
	}

	sCfg := server.Config{
		AuthKey: "abcdefg",
		KCP: server.KCPConfig{
			Listen:       "0.0.0.0:1001",
			Remote:       targetAddr,
			MTU:          1200,
			SndWnd:       2048,
			RcvWnd:       256,
			DataShard:    30,
			ParityShard:  15,
			DSCP:         46,
			AckNodelay:   true,
			NoDelay:      1,
			Interval:     10,
			Resend:       2,
			NoCongestion: 1,
			SockBuf:      16777217,
			StreamBuf:    4194304,
		},
		QUIC: server.QUICConfig{},
		TCP:  server.TCPConfig{},
	}

	err := startKCPServer(sCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startClient("kcp", cCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startTargetServer(targetAddr)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 2)

	err = echo(localProxyAddr)
	if err != nil {
		t.Fatal(err)
	}
}

func startKCPServer(config server.Config) error {
	aead, err := util.CreateAesGcmAead(util.StringToAesKey(config.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}
	s := server.NewKCPServer(config.KCP, aead)
	err = s.Run()
	if err != nil {
		return err
	}
	return nil
}

func Test_QUIC(t *testing.T) {
	targetAddr := "127.0.0.1:3000"
	localProxyAddr := "127.0.0.1:1000"

	cCfg := client.Config{
		AuthKey: "abcdefg",
		ConnNum: 10,
		QUIC: client.QUICConfig{
			Listen: localProxyAddr,
			Remote: "127.0.0.1:1001",
		},
		TCP: client.TCPConfig{},
	}

	sCfg := server.Config{
		AuthKey: "abcdefg",
		QUIC: server.QUICConfig{
			Listen: "0.0.0.0:1001",
			Remote: targetAddr,
		},
		TCP: server.TCPConfig{},
	}

	err := startQUICServer(sCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startClient("quic", cCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startTargetServer(targetAddr)
	if err != nil {
		t.Fatal(err)
	}

	err = echo(localProxyAddr)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_TCP(t *testing.T) {
	targetAddr := "127.0.0.1:3000"
	localProxyAddr := "127.0.0.1:1000"

	cCfg := client.Config{
		AuthKey: "abcdefg",
		ConnNum: 10,
		QUIC:    client.QUICConfig{},
		TCP: client.TCPConfig{
			Listen: localProxyAddr,
			Remote: "127.0.0.1:1001",
		},
	}

	sCfg := server.Config{
		AuthKey: "abcdefg",
		QUIC:    server.QUICConfig{},
		TCP: server.TCPConfig{
			Listen: "0.0.0.0:1001",
			Remote: targetAddr,
		},
	}

	err := startClient("tcp", cCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startTCPServer(sCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startTargetServer(targetAddr)
	if err != nil {
		t.Fatal(err)
	}

	err = echo(localProxyAddr)
	if err != nil {
		t.Fatal(err)
	}
}

func startClient(typ string, config client.Config) error {
	aead, err := util.CreateAesGcmAead(util.StringToAesKey(config.AuthKey, 32))
	if err != nil {
		return err
	}

	c, err := client.NewTunClient(typ, config, aead)
	if err != nil {
		return err
	}
	err = c.Run()
	if err != nil {
		return err
	}
	return nil
}

func startQUICServer(config server.Config) error {
	aead, err := util.CreateAesGcmAead(util.StringToAesKey(config.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}
	s := server.NewQUICServer(config.QUIC, aead)
	err = s.Run()
	if err != nil {
		return err
	}
	return nil
}

func startTCPServer(config server.Config) error {
	aead, err := util.CreateAesGcmAead(util.StringToAesKey(config.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}
	s := server.NewTCPServer(config.TCP, aead)
	err = s.Run()
	if err != nil {
		return err
	}
	return nil
}

func startTargetServer(targetAddr string) error {
	listener, err := net.Listen("tcp", targetAddr)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, 65535)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println(err)
			}
			conn.Write(buf[0:n])
		}
	}()
	return nil
}

func echo(clientAddr string) error {
	conn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		return err
	}

	//sendStr := "hello"
	t := time.Now().UnixNano()
	s := fmt.Sprintf("%v", t)

	fmt.Println("sendTime:", time.Now().UnixNano()/1000000)
	_, err = conn.Write([]byte(s))
	if err != nil {
		return err
	}

	//conn.SetReadDeadline(time.Now().Add(time.Second * 20))
	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
	}

	retStr := string(buf[0:n])

	if retStr != s {
		return fmt.Errorf("echo send:%s receive:%s", s, buf[0:n])
	}

	fmt.Println("receiveTime:", time.Now().UnixNano()/1000000)
	latency := (time.Now().UnixNano() - t)
	fmt.Println("latency:", latency/1000000)
	return nil
}
