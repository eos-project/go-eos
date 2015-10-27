package udp

import (
	"github.com/eos-project/go-eos/model"
	"github.com/eos-project/go-eos/server"
)

type Config struct {
	Address    string
	PacketSize int
	BufferSize int

	StatServe        func()
	StatErrorConnect func()
	StatErrorParse   func()
	StatErrorAuth    func()

	Send         server.Listener
	Authenticate func(model.Packet) error
}
