package ws

import "github.com/eos-project/go-eos/server"

type Config struct {
	Address string

	Dispatcher *server.Dispatcher
	//	Authenticate func(model.Packet) error
}
