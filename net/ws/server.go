package ws

import (
	"github.com/gorilla/websocket"
	"github.com/gotterdemarung/go-log/log"
	"net"
	"net/http"
)

var httpLog = log.Context.WithTags("eos", "udp")

func StartListening(c Config) (func(), error) {
	socket, err := net.Listen("tcp", c.Address)
	if err != nil {
		return nil, httpLog.Fail(err)
	}

	httpLog.Context["addr"] = c.Address
	httpLog.Info("Starting HTTP server on :addr")
	mux := http.NewServeMux()
	mux.HandleFunc("/", getWsHandler(&c))
	srv := &http.Server{Handler: mux}
	go srv.Serve(socket)

	return func() {
		socket.Close()
	}, nil
}

func getWsHandler(c *Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			return
		}

		StartWsConnector(ws, c.Dispatcher)
	}
}
