package main

import (
	"github.com/tsheasha/relayd/config"
	"github.com/tsheasha/relayd/forwarder"
)

func startForwarders(c config.Config) (forwarders []forwarder.Forwarder) {
	log.Info("Starting forwarders...")
	for name, config := range c.Forwarders {
		forwarders = append(forwarders, startForwarder(name, c, config))
	}
	return
}

func startForwarder(name string, globalConfig config.Config, instanceConfig map[string]interface{}) forwarder.Forwarder {
	log.Info("Starting forwarder ", name)
	f := forwarder.New(name)
	if f == nil {
		return nil
	}

	// now apply the forwarder level configs
	f.Configure(instanceConfig)

	// now run a channel for each listener
	f.InitListeners(globalConfig)

	go f.Run()
	return f
}
