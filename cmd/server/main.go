package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/stunnel/logconfig"
	"github.com/0990/stunnel/server"
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var confFile = flag.String("c", "stserver.json", "config file")

func main() {
	logconfig.InitLogrus("stserver", 10, logrus.WarnLevel)

	flag.Parse()

	file, err := os.Open(*confFile)
	if err != nil {
		logrus.Fatalln(err)
	}

	var config server.Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.Info("config:", config)

	aead, err := util.CreateAesGcmAead(util.StringToAesKey(config.AuthKey, 32))
	if err != nil {
		logrus.Fatalln(err)
	}

	if config.KCP.Listen != "" {
		s := server.NewKCPServer(config.KCP, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		err := s.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	if config.QUIC.Listen != "" {
		s := server.NewQUICServer(config.QUIC, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		err := s.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	if config.TCP.Listen != "" {
		s := server.NewTCPServer(config.TCP, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		err := s.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	if config.RawUDP.Listen != "" {
		c := server.NewRawUDP(config.RawUDP, aead)
		err = c.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("quit,Got signal:", s)
}
