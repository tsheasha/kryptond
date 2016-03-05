package listener

import (
	"relayd/config"

	l "github.com/Sirupsen/logrus"
)

const (
	// DefaultMaxBufferSize indicates the read buffer
	// for the socket on the OS level
	DefaultMaxBufferSize = 16777216

	// DefaultMaxMsgSize indicates the maximum message
	// size to be received by a listener
	DefaultMaxMsgSize = 65536
)

var defaultLog = l.WithFields(l.Fields{"app": "relayd", "pkg": "listener"})

// Listener defines the interface of a generic listener.
type Listener interface {
	Listen()
	Configure(map[string]interface{})

	// taken care of by the base class
	Channel() chan []byte
	MaxMsgSize() int
	Name() string
	ReadBuffer() int
}

var listenerConstructs map[string]func(chan []byte, *l.Entry) Listener

// RegisterListener composes a map of listener names -> factory functions
func RegisterListener(name string, f func(chan []byte, *l.Entry) Listener) {
	if listenerConstructs == nil {
		listenerConstructs = make(map[string]func(chan []byte, *l.Entry) Listener)
	}
	listenerConstructs[name] = f
}

// New creates a new Listener based on the requested listener name.
func New(name string) Listener {
	var listener Listener

	channel := make(chan []byte)
	listenerLog := defaultLog.WithFields(l.Fields{"listener": name})

	if f, exists := listenerConstructs[name]; exists {
		listener = f(channel, listenerLog)
	} else {
		defaultLog.Error("Cannot create listener: ", name)
		return nil
	}

	return listener
}

type baseListener struct {
	// fulfill most of the rote parts of the listener interface
	channel    chan []byte
	maxMsgSize int
	name       string
	readBuffer int

	// intentionally exported
	log *l.Entry
}

func (l *baseListener) configureCommonParams(configMap map[string]interface{}) {
	l.readBuffer = DefaultMaxBufferSize
	if b, exists := configMap["readBuffer"]; exists {
		l.readBuffer = config.GetAsInt(b, DefaultMaxBufferSize)
	}

	l.maxMsgSize = DefaultMaxMsgSize
	if m, exists := configMap["maxMsgSize"]; exists {
		l.maxMsgSize = config.GetAsInt(m, DefaultMaxMsgSize)
	}
}

// Channel : the channel on which the listener should send messages
func (l baseListener) Channel() chan []byte {
	return l.channel
}

// ReadBuffer : the OS level protocol socket buffer
func (l baseListener) ReadBuffer() int {
	return l.readBuffer
}

// MaxMsgSize : max size of incoming message
func (l baseListener) MaxMsgSize() int {
	return l.maxMsgSize
}

// Name : the name of the listener
func (l baseListener) Name() string {
	return l.name
}

// String returns the listener name in printable format.
func (l baseListener) String() string {
	return l.Name() + "Listener"
}
