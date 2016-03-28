package main

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/davecheney/profile"
	"github.com/tsheasha/relayd/config"
	"github.com/tsheasha/relayd/internalserver"
)

const (
	name    = "relayd"
	version = "0.0.1"
	desc    = "Content-agnostic Message relay daemon"
)

var log = logrus.WithFields(logrus.Fields{"app": "relayd"})

func initLogrus(ctx *cli.Context) {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: time.RFC822,
		FullTimestamp:   true,
	})

	if level, err := logrus.ParseLevel(ctx.String("log_level")); err == nil {
		logrus.SetLevel(level)
	} else {
		log.Error(err)
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Usage = desc
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/relayd.conf",
			Usage: "JSON formatted configuration file",
		},
		cli.StringFlag{
			Name:  "log_level, l",
			Value: "info",
			Usage: "Logging level (debug, info, warn, error, fatal, panic)",
		},
		cli.BoolFlag{
			Name:  "profile",
			Usage: "Enable profiling",
		},
	}
	app.Action = start

	app.Run(os.Args)
}

func start(ctx *cli.Context) {
	if ctx.Bool("profile") {
		pcfg := profile.Config{
			CPUProfile:   true,
			MemProfile:   true,
			BlockProfile: true,
			ProfilePath:  ".",
		}
		p := profile.Start(&pcfg)
		defer p.Stop()
	}
	quit := make(chan bool)
	initLogrus(ctx)
	log.Info("Starting relayd...")

	c, err := config.ReadConfig(ctx.String("config"))
	if err != nil {
		return
	}
	listeners := startListeners(c)
	forwarders := startForwarders(c)

	internalServer := internalserver.New(c, &forwarders)
	go internalServer.Run()

	readFromListeners(listeners, forwarders)

	<-quit
}
