package listener

import (
	"bufio"
	"net"
	"strings"
	"time"

	l "github.com/Sirupsen/logrus"
)

const (
	// DefaultTCPListenerPort is the default port
	// to listen on for incoming TCP traffic
	DefaultTCPListenerPort = "19191"
)

// TCP listener type
type TCP struct {
	baseListener
	port string
}

func init() {
	RegisterListener("TCP", newTCP)
}

// newTCP creates a new TCP listener.
func newTCP(channel chan []byte, log *l.Entry) Listener {
	t := new(TCP)

	t.log = log
	t.channel = channel

	t.name = "TCP"
	t.port = DefaultTCPListenerPort
	return t
}

// Configure the listener
func (t *TCP) Configure(configMap map[string]interface{}) {
	if port, exists := configMap["port"]; exists {
		t.port = port.(string)
	}
	t.configureCommonParams(configMap)
}

// Listen passes incoming traffic to the channel to be picked up
// by forwarder
func (t *TCP) Listen() {
	addr, err := net.ResolveTCPAddr("tcp", ":"+t.port)

	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		t.log.Fatal("Cannot listen on socket", err)
	}

	// figure out the port bind for Port()
	t.port = strings.Split(l.Addr().String(), ":")[1]

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			t.log.Fatal(err)
		}

		go t.readMessage(conn)
	}
}

// readMessage reads from the connection
func (t *TCP) readMessage(conn *net.TCPConn) {
	defer conn.Close()
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second)
	conn.SetReadBuffer(t.ReadBuffer())

	reader := bufio.NewReader(conn)
	t.log.Info("Connection started: ", conn.RemoteAddr())

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.log.Warn("Error while reading message", err)
			break
		}
		t.log.Debug("Read: ", string(line))
		t.Channel() <- line
	}
	t.log.Info("Connection closed: ", conn.RemoteAddr())
}
