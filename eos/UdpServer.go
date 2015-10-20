package eos

import (
	"github.com/gotterdemarung/go-log/log"
	"net"
	"strings"
)

var udpLog = log.Context.WithTags("eos", "udp")

type UdpServer struct {
	config  UdpServerConfiguration
	address *net.UDPAddr
	socket  *net.UDPConn
	running bool
}

type UdpServerConfiguration struct {
	Address    string
	PacketSize int
	BufferSize int

	StatServe        func()
	StatErrorConnect func()
	StatErrorParse   func()
	StatErrorAuth    func()

	ParseKey     func(string) (*EosKey, error)
	Send         func(Message)
	Authenticate func(Packet) error
}

func NewUdpServer(cnf UdpServerConfiguration) (*UdpServer, error) {
	var err error
	udpLog.Info("Initializing UDP listener")

	s := UdpServer{config: cnf}
	s.address, err = net.ResolveUDPAddr("udp", s.config.Address)
	if err != nil {
		return nil, udpLog.Fail(err)
	}

	// Filling defaults
	if s.config.PacketSize == 0 {
		s.config.PacketSize = 1024 * 8
	}
	if s.config.BufferSize == 0 {
		s.config.BufferSize = 1024 * 1024 * 4
	}
	if s.config.StatErrorConnect == nil {
		s.config.StatErrorConnect = func() {}
	}
	if s.config.StatErrorParse == nil {
		s.config.StatErrorParse = func() {}
	}
	if s.config.StatErrorAuth == nil {
		s.config.StatErrorAuth = func() {}
	}

	return &s, nil
}

func (u *UdpServer) Start() error {
	var err error

	udpLog.Infoc(
		"Lisening UDP at :addr, max packet size :mpsize, incoming buffer size :ibsize",
		map[string]interface{}{
			"addr":   u.config.Address,
			"mpsize": u.config.PacketSize,
			"ibsize": u.config.BufferSize,
		},
	)

	u.socket, err = net.ListenUDP("udp", u.address)
	if err != nil {
		udpLog.Fail(err)
		return err
	}

	u.running = true
	go func() {
		for u.running {
			buf := make([]byte, u.config.PacketSize)
			rlen, _, err := u.socket.ReadFromUDP(buf)
			u.config.StatServe()
			if err != nil {
				u.config.StatErrorConnect()
			} else {
				go u.accept(string(buf[0:rlen]))
			}
		}
	}()

	return nil
}

func (u *UdpServer) Stop() {
	udpLog.Info("Closing UDP listener")
	u.running = false
	u.socket.Close()
}

func (u *UdpServer) accept(packet string) {
	chunks := strings.Split(packet, "\n")

	if len(chunks) < 4 {
		u.config.StatErrorParse()
		return
	}

	nonce := chunks[0]
	signature := chunks[1]
	keyString := chunks[2]

	// Resolving key
	key, err := u.config.ParseKey(keyString)
	if err != nil {
		u.config.StatErrorParse()
		return
	}

	// Build packet
	msg := Message{
		*key,
		strings.TrimSpace(strings.Join(chunks[3:], "\n")),
	}
	pkt := Packet{
		msg,
		nonce,
		signature,
	}

	// Authenticate
	err = u.config.Authenticate(pkt)
	if err != nil {
		u.config.StatErrorAuth()
		return
	}

	// Send
	u.config.Send(msg)
}
