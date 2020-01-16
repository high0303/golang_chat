package lib

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"time"
)

func SLog(format string, a ...interface{}) {
	// Get name of calling function
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	logmessage := fmt.Sprintf(format, a...)
	log.Printf("%s: %s", caller, logmessage)
}

var FederationPort string = "10001"

type FederationUser struct {
	Username string
	Domain   string
}

func FederationUserFromString(useraddress string) FederationUser {
	address_components := strings.Split(useraddress, "@")
	if len(address_components) != 2 {
		// TODO Error!
		panic("Don't have 2 address components!")
	}
	fu := FederationUser{Username: address_components[0], Domain: address_components[1]}
	return fu
}

type FederationMessage struct {
	Payload            []byte
	Receiver           string
	Sender             string
	MessageId          string
	ValidUntil         string
	ImmediateOnly      bool
	Priority           int
	Notification       int
	RepeatNotification bool
}

// --------------------------------------------
// Federation Connection
// --------------------------------------------

type FederationConnection struct {
	Connection     net.Conn
	PayloadChannel chan<- *FederationMessage
	Domain         string
	LastActivity   time.Time
}

func FederationConnectionCreate(connection net.Conn, domain string) FederationConnection {
	fc := FederationConnection{
		Connection: connection,
		Domain:     domain,
	}

	fc.PayloadChannel = make(chan *FederationMessage, 1024)

	fc.UpdateLastActivity()

	return fc
}

func (fc FederationConnection) Send(data []byte) {
	fc.Connection.Write(data)
	fc.UpdateLastActivity()
}

func (fc FederationConnection) UpdateLastActivity() {
	fc.LastActivity = time.Now()
}

// --------------------------------------------
// Federation Manager
// --------------------------------------------

type FederationManager struct {
	Connections map[string]FederationConnection
	Domain      string
}

func FederationManagerCreate(domain string) FederationManager {
	fm := FederationManager{
		Connections: make(map[string]FederationConnection),
		Domain:      domain,
	}

	return fm
}

func (fm FederationManager) ListenerService() {
	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", FederationPort))

	if err != nil {
		panic("Failed to start listener service")
	}

	SLog("listening")

	// Endless listener loop
	for {
		// TODO some kind of handshaking...
		conn, err := listener.Accept()
		SLog("Accepted %s", conn.RemoteAddr())
		if err != nil {
			panic("Failed to accept connection")
		}

		fc := FederationConnectionCreate(conn, "sealed2.ch")

		fm.RegisterConnection(fc)

		defer conn.Close()
	}
}

func (fm FederationManager) InitiateConnection(host string, port string) FederationConnection {
	address := net.JoinHostPort(host, port)
	SLog("Initiating connection to: %s", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		// TODO handle error
		panic("Failed to connect!")
	}

	fc := FederationConnection{Connection: conn}
	fc.UpdateLastActivity()

	fm.Connections[host] = fc

	return fc
}

func (fm FederationManager) InitiateOrGetConnection(host string, port string) FederationConnection {
	if _, ok := fm.Connections[host]; ok {
		return fm.Connections[host]
	}
	return fm.InitiateConnection(host, port)
}

func (fm FederationManager) RegisterConnection(fc FederationConnection) {
	fm.Connections[fc.Domain] = fc
}

func (fm FederationManager) SendToUser(user string, container InterconnectMessageContainer) {
	fu := FederationUserFromString(user)

	// Check if the user is a local user
	if fu.Domain == fm.Domain {
		// Local user

	} else { /// Local user
		// External user
		connection := fm.InitiateOrGetConnection(fu.Domain, FederationPort)
		connection.Send(container.Frame)
	} /// External user
}
