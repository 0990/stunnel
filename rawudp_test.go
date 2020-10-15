package stunnel

import (
	"fmt"
	"github.com/0990/stunnel/client"
	"github.com/0990/stunnel/server"
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
	"net"
	"testing"
	"time"
)

func Test_RawUDP(t *testing.T) {
	targetAddr := "127.0.0.1:2000"
	localProxyAddr := "127.0.0.1:1000"

	cCfg := client.Config{
		AuthKey: "abcdefg",
		ConnNum: 10,
		QUIC:    client.QUICConfig{},
		TCP:     client.TCPConfig{},
		RawUDP: client.RawUDPConfig{
			Listen:  localProxyAddr,
			Remote:  "127.0.0.1:1001",
			Timeout: 30,
		},
	}

	sCfg := server.Config{
		AuthKey: "abcdefg",
		QUIC:    server.QUICConfig{},
		TCP:     server.TCPConfig{},
		RawUDP: server.RawUDPConfig{
			Listen:  "0.0.0.0:1001",
			Remote:  targetAddr,
			Timeout: 30,
		},
	}

	err := startRawUDPServer(sCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startRawUDPClient(cCfg)
	if err != nil {
		t.Fatal(err)
	}

	err = startUDPTargetServer(targetAddr)
	if err != nil {
		t.Fatal(err)
	}

	err = echoUDP(localProxyAddr)
	if err != nil {
		t.Fatal(err)
	}
}

func startRawUDPClient(cfg client.Config) error {
	aead, err := util.CreateAesGcmAead(util.StringToAesKey(cfg.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}
	c := client.NewRawUDPClient(cfg.RawUDP, aead)
	return c.Run()
}

func startRawUDPServer(cfg server.Config) error {
	aead, err := util.CreateAesGcmAead(util.StringToAesKey(cfg.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}
	c := server.NewRawUDP(cfg.RawUDP, aead)
	return c.Run()
}

func echoUDP(clientAddr string) error {
	addr, err := net.ResolveUDPAddr("udp", clientAddr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	sendStr := "hello"
	_, err = conn.Write([]byte(sendStr))
	if err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 20))
	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	retStr := string(buf[0:n])

	if retStr != sendStr {
		return fmt.Errorf("echo send:%s receive:%s", sendStr, buf[0:n])
	}
	return nil
}

func startUDPTargetServer(targetAddr string) error {
	addr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return err
	}
	listen, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	go func() {
		for {
			var data [1024]byte
			n, addr, err := listen.ReadFromUDP(data[:])
			if err != nil {
				fmt.Println(err)
				break
			}

			fmt.Printf("addr:%v data:%v count:%v\n", addr, string(data[:n]), n)
			_, err = listen.WriteToUDP(data[:n], addr)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()
	return nil
}
