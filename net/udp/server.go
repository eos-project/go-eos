package udp

import (
	"github.com/eos-project/go-eos/encoding/native"
	"github.com/gotterdemarung/go-log/log"
	"net"
)

var udpLog = log.Context.WithTags("eos", "udp")

func StartListening(c Config) (func(), error) {
	var err error
	udpLog.Info("Initializing UDP listener")

	address, err := net.ResolveUDPAddr("udp", c.Address)
	if err != nil {
		return nil, udpLog.Fail(err)
	}

	// Filling defaults
	if c.PacketSize == 0 {
		c.PacketSize = 1024 * 8
	}
	if c.BufferSize == 0 {
		c.BufferSize = 1024 * 1024 * 4
	}
	if c.StatErrorConnect == nil {
		c.StatErrorConnect = func() {}
	}
	if c.StatErrorParse == nil {
		c.StatErrorParse = func() {}
	}
	if c.StatErrorAuth == nil {
		c.StatErrorAuth = func() {}
	}

	udpLog.Context["addr"] = c.Address
	udpLog.Context["mpsize"] = c.PacketSize
	udpLog.Context["ibsize"] = c.BufferSize
	udpLog.Info("Lisening UDP at :addr, max packet size :mpsize, incoming buffer size :ibsize")

	socket, err := net.ListenUDP("udp", address)
	if err != nil {
		return nil, udpLog.Fail(err)
	}

	running := true

	// Listener
	go func() {
		for running {
			buf := make([]byte, c.PacketSize)
			rlen, _, err := socket.ReadFromUDP(buf)
			c.StatServe()
			if err != nil {
				c.StatErrorConnect()
			} else {
				go accept(buf[0:rlen], &c)
			}
		}
		socket.Close()
	}()

	// Returns STOP-function
	return func() {
		running = false
	}, nil
}

func accept(data []byte, c *Config) {
	pkt, err := native.UnmarshalPacket(data)
	if err != nil {
		// Parse error
		c.StatErrorParse()
		return
	}

	err = c.Authenticate(*pkt)
	if err != nil {
		// Auth error
		c.StatErrorAuth()
		return
	}

	// OK
	c.Send(pkt.Message)
}
