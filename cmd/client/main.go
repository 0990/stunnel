package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/stunnel/client"
	"github.com/0990/stunnel/logconfig"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var confFile = flag.String("c", "stclient.json", "config file")

func main() {
	flag.Parse()

	file, err := os.Open(*confFile)
	if err != nil {
		logrus.Fatalln(err)
	}

	var cfg client.Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.Info("config:", cfg)

	if len(cfg.Tunnels) == 0 {
		logrus.Fatalln("no tunnels")
	}

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.Fatalln(err)
	}

	logconfig.InitLogrus("stclient", 10, level)

	for _, tunCfg := range cfg.Tunnels {
		p := client.New(tunCfg)
		err = p.Run()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("quit,Got signal:", s)
}
