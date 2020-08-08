package stunnel

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func Test_STunnel(t *testing.T) {
	targetAddr := "127.0.0.1:1234"
	remoteAddr := "127.0.0.1:1233"
	localAddr := "127.0.0.1:1232"
	err := startTargetServer(targetAddr)
	if err != nil {
		t.Fatal(err)
	}

	err = startServer(SConfig{
		Listen: remoteAddr,
		Target: targetAddr,
	})

	if err != nil {
		t.Fatal(err)
	}

	err = startClient(CConfig{
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = echo(localAddr)
	if err != nil {
		t.Fatal(err)
	}

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

func startClient(config CConfig) error {
	c := NewClient(config)
	err := c.Run()
	if err != nil {
		return err
	}
	return nil
}

func startServer(config SConfig) error {
	s := newServer(config)
	err := s.Run()
	if err != nil {
		return err
	}
	return nil
}

func echo(clientAddr string) error {
	conn, err := net.Dial("tcp", clientAddr)
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
		fmt.Println(err)
	}

	if string(buf[0:n]) != sendStr {
		return fmt.Errorf("echo send:%s receive:%s", sendStr, buf[0:n])
	}
	return nil
}
