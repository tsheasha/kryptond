package forwarder

import (
	"sync/atomic"

	l "github.com/Sirupsen/logrus"
	"github.com/tsheasha/relayd/config"
)

// Some sane values to default things to
const (
	DefaultBufferSize        = 100
	DefaultKeepAliveInterval = 30
)

var defaultLog = l.WithFields(l.Fields{"app": "relayd", "pkg": "forwarder"})

var forwarderConstructs map[string]func(int, *l.Entry) Forwarder

// RegisterForwarder takes forwarder name and constructor function and returns a forwarder
func RegisterForwarder(name string, f func(int, *l.Entry) Forwarder) {
	if forwarderConstructs == nil {
		forwarderConstructs = make(map[string]func(int, *l.Entry) Forwarder)
	}
	forwarderConstructs[name] = f
}

// New creates a new Forwarder based on the requested forwarder name.
func New(name string) Forwarder {
	forwarderLog := defaultLog.WithFields(l.Fields{"forwarder": name})

	if f, exists := forwarderConstructs[name]; exists {
		return f(DefaultBufferSize, forwarderLog)
	}

	defaultLog.Error("Cannot create forwarder ", name)
	return nil
}

// InternalMetrics holds the key:value pairs for counters/gauges
type InternalMetrics struct {
	Counters map[string]float64
	Gauges   map[string]float64
}

// NewInternalMetrics initializes the internal components of InternalMetrics
func NewInternalMetrics() *InternalMetrics {
	inst := new(InternalMetrics)
	inst.Counters = make(map[string]float64)
	inst.Gauges = make(map[string]float64)
	return inst
}

// Forwarder defines the interface of a generic forwarder.
type Forwarder interface {
	Run()
	Configure(map[string]interface{})
	InitListeners(config.Config)

	// InternalMetrics is to publish a set of values
	// that are relevant to the forwarder itself.
	InternalMetrics() InternalMetrics

	// taken care of by the base
	Name() string
	String() string

	ListenerChannels() map[string]chan []byte
	SetListenerChannels(map[string]chan []byte)

	MaxBufferSize() int
	SetMaxBufferSize(int)

	KeepAliveInterval() int
	SetKeepAliveInterval(int)
}

// BaseForwarder is class to handle the boiler plate parts of the forwarders
type BaseForwarder struct {
	listenerChannels map[string]chan []byte
	name             string
	log              *l.Entry

	maxBufferSize int

	// for keepalive
	keepAliveInterval int

	totalEmissions uint64
	msgsSent       uint64
	msgsDropped    uint64
}

// SetMaxBufferSize : set the buffer size
func (base *BaseForwarder) SetMaxBufferSize(size int) {
	base.maxBufferSize = size
}

// ListenerChannels : the channels to forwarders listens for messages on
func (base BaseForwarder) ListenerChannels() map[string]chan []byte {
	return base.listenerChannels
}

// SetListenerChannels : the channels to forwarder listens for messages on
func (base *BaseForwarder) SetListenerChannels(c map[string]chan []byte) {
	base.listenerChannels = make(map[string]chan []byte)
	for name, channel := range c {
		base.listenerChannels[name] = channel
	}
}

// Name : the name of the forwarder
func (base BaseForwarder) Name() string {
	return base.name
}

// MaxBufferSize : the maximum number of messages to be in the circular buffer
func (base BaseForwarder) MaxBufferSize() int {
	return base.maxBufferSize
}

// SetKeepAliveInterval : Set keep alive interval
func (base *BaseForwarder) SetKeepAliveInterval(value int) {
	base.keepAliveInterval = value
}

// InitListeners - initiate channels for listeners
func (base *BaseForwarder) InitListeners(globalConfig config.Config) {
	lietenerChannels := make(map[string]chan []byte)
	for name := range globalConfig.Listeners {
		lietenerChannels[name] = make(chan []byte, base.MaxBufferSize())
	}
	base.SetListenerChannels(lietenerChannels)
}

// KeepAliveInterval - return keep alive interval
func (base BaseForwarder) KeepAliveInterval() int {
	return base.keepAliveInterval
}

// String returns the forwarder name in a printable format.
func (base BaseForwarder) String() string {
	return base.name + "Forwarder"
}

// InternalMetrics : Returns the internal metrics that are being collected by this forwarder
func (base BaseForwarder) InternalMetrics() InternalMetrics {
	counters := map[string]float64{
		"totalEmissions": float64(base.totalEmissions),
		"msgsDropped":    float64(base.msgsDropped),
		"msgsSent":       float64(base.msgsSent),
	}

	return InternalMetrics{
		Counters: counters,
	}
}

// configureCommonParams will extract the common parameters that are used and set them in the forwarder
func (base *BaseForwarder) configureCommonParams(configMap map[string]interface{}) {

	if asInterface, exists := configMap["max_buffer_size"]; exists {
		base.maxBufferSize = config.GetAsInt(asInterface, DefaultBufferSize)
	}

	if asInterface, exists := configMap["keepAliveInterval"]; exists {
		keepAliveInterval := config.GetAsInt(asInterface, DefaultKeepAliveInterval)
		base.SetKeepAliveInterval(keepAliveInterval)
	}
}

func (base *BaseForwarder) run(emitFunc func([]byte) bool) {
	for k := range base.ListenerChannels() {
		go base.listenForMsgs(emitFunc, base.ListenerChannels()[k])
	}
}

func (base *BaseForwarder) listenForMsgs(
	emitFunc func([]byte) bool,
	c <-chan []byte) {

	for incomingMsg := range c {
		base.log.Debug(base.Name(), " msg: ", string(incomingMsg))
		result := emitFunc(incomingMsg)

		if result {
			atomic.AddUint64(&base.msgsSent, 1)
			base.log.Debug("Relay Successful")
		} else {
			base.log.Debug("Relay Failed")
			atomic.AddUint64(&base.msgsDropped, 1)
		}
	}
}
