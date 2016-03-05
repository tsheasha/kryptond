package forwarder

import (
	"net"

	l "github.com/Sirupsen/logrus"
)

func init() {
	RegisterForwarder("UDP", newUDP)
}

// UDP forwarder
type UDP struct {
	BaseForwarder
	conn   *net.UDPConn
	server string
	port   string
}

// newUDP returns a new UDP forwarder
func newUDP(
	initialBufferSize int,
	log *l.Entry) Forwarder {

	u := new(UDP)
	u.name = "UDP"

	u.maxBufferSize = initialBufferSize
	u.log = log
	return u
}

// Configure the UDP forwader
func (u *UDP) Configure(configMap map[string]interface{}) {
	if server, exists := configMap["server"]; exists {
		u.server = server.(string)
	} else {
		u.log.Error("There was no server specified, there won't be any emissions")
	}

	if port, exists := configMap["port"]; exists {
		u.port = port.(string)
	} else {
		u.log.Error("There was no port specified , there won't be any emissions")
	}
	u.configureCommonParams(configMap)
}

// Run runs the forwarder main loop
func (u *UDP) Run() {
	addr, err := net.ResolveUDPAddr("udp", u.server+":"+u.port)
	if err != nil {
		u.log.Error("Could not resolve remote UDP address")
	}

	u.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		u.log.Error("Could not connect to remote UDP host")
	}

	u.run(u.emitMsg)
}

func (u *UDP) emitMsg(m []byte) bool {

	_, err := u.conn.Write(m)
	if err != nil {
		u.log.Error("Failed to send message to UDP endpoint")
		return false
	}

	return true
}
