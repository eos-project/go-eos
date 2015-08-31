package eos

import (
	"net/http"
	"github.com/gotterdemarung/go-log/log"
	"github.com/gorilla/websocket"
)

var httpLog = log.Context.WithTags("eos", "http")

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}

	// Building and starting new websocket listener
	l := WebsocketListener{
		ch: make(chan Message),
		ws: ws,
	}

	l.accept()
}


type WebsocketListener struct {
	ch chan Message
	ws *websocket.Conn
}

func (l *WebsocketListener) OnMessage(message Message) {
	l.ch <- message
}

func (l *WebsocketListener) accept() {
	for {
		_, command, err := l.ws.ReadMessage()
		if err != nil {
			break;
		}

		httpLog.Trace(string(command))
	}

	l.ws.Close()
}

type HttpServer struct {
	addr	string
	server 	*http.Server
}

func NewHttpServer(addr string) *HttpServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/", wsHandler)
	server := &http.Server{Addr: addr, Handler: mux}

	hs := HttpServer{
		server: server,
	}

	return &hs
}

func (h *HttpServer) Start() {
	go func() {
		httpLog.Infoc("Starting HTTP server on :addr", map[string]interface{}{"addr": h.server.Addr})
		err := h.server.ListenAndServe()
		if err != nil {
			httpLog.Fail(err)
			panic(err)
		}
	}()
}