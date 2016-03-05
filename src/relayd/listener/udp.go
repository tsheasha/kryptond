package listener

import (
	"net"

	l "github.com/Sirupsen/logrus"
)

const (
	// DefaultUDPListenerPort is the default port
	// to listen on for incoming UDP traffic
	DefaultUDPListenerPort = "19192"
)

// UDP listener type
type UDP struct {
	baseListener
	port string
}

func init() {
	RegisterListener("UDP", newUDP)
}

// newUDP creates a new UDP listener.
func newUDP(channel chan []byte, log *l.Entry) Listener {
	u := new(UDP)

	u.log = log
	u.channel = channel

	u.name = "UDP"
	u.port = DefaultUDPListenerPort
	return u
}

// Configure the listener
func (u *UDP) Configure(configMap map[string]interface{}) {
	if port, exists := configMap["port"]; exists {
		u.port = port.(string)
	}

	u.configureCommonParams(configMap)
}

// Listen passes incoming traffic to the channel to be picked up
// by forwarder
func (u *UDP) Listen() {
	addr, err := net.ResolveUDPAddr("udp", ":"+u.port)

	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		u.log.Fatal("Cannot listen on socket", err)
	}

	defer conn.Close()

	conn.SetReadBuffer(u.ReadBuffer())
	line := make([]byte, u.MaxMsgSize())

	for {
		n, err := conn.Read(line)
		if err != nil {
			u.log.Warn("Error while reading message: ", err)
			break
		}
		u.log.Debug("Read: ", string(line[0:n]))
		u.Channel() <- line[0:n]
	}
}
