package main

import (
	"relayd/config"
	"relayd/forwarder"
	"relayd/listener"
)

func startListeners(c config.Config) (listeners []listener.Listener) {
	log.Info("Starting listeners...")

	for name, conf := range c.Listeners {
		l := startListener(name, c, conf)
		if l != nil {
			listeners = append(listeners, l)
		}
	}
	return
}

func startListener(name string, globalConfig config.Config, instanceConfig map[string]interface{}) listener.Listener {
	log.Debug("Starting listener ", name)
	l := listener.New(name)
	if l == nil {
		return nil
	}

	// apply the instance configs
	l.Configure(instanceConfig)

	log.Info("Running ", l)
	go l.Listen()

	return l
}

func readFromListeners(listeners []listener.Listener, forwarders []forwarder.Forwarder) {
	for i := range listeners {
		go readFromListener(listeners[i], forwarders)
	}
}

func readFromListener(l listener.Listener, forwarders []forwarder.Forwarder) {
	for msg := range l.Channel() {
		for i := range forwarders {
			if _, exists := forwarders[i].ListenerChannels()[l.Name()]; exists {
				forwarders[i].ListenerChannels()[l.Name()] <- msg
			}
		}
	}
}
