package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/stunnel/client"
	"github.com/0990/stunnel/logconfig"
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var confFile = flag.String("c", "stclient.json", "config file")

func main() {
	logconfig.InitLogrus("stclient", 10, logrus.ErrorLevel)

	flag.Parse()

	file, err := os.Open(*confFile)
	if err != nil {
		logrus.Fatalln(err)
	}

	var config client.Config
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
		c, err := client.New("kcp", config, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		err = c.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	if config.QUIC.Listen != "" {
		c, err := client.New("quic", config, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		err = c.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	if config.TCP.Listen != "" {
		c, err := client.New("tcp", config, aead)
		if err != nil {
			logrus.Fatalln(err)
		}
		err = c.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	if config.RawUDP.Listen != "" {
		c := client.NewRawUDP(config.RawUDP, aead)
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
