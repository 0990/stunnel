package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/stunnel"
	"github.com/0990/stunnel/logconfig"
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

	var config stunnel.CConfig
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.Info("config:", config)

	client := stunnel.NewClient(config)
	err = client.Run()
	if err != nil {
		logrus.Fatalln(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("quit,Got signal:", s)
}