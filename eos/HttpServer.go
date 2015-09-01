package eos

import (
	"fmt"
	"strings"
	"net"
	"net/http"
	"encoding/json"
	"crypto/rand"
	"github.com/gotterdemarung/go-log/log"
	"github.com/gorilla/websocket"
	"golang.org/x/tools/go/buildutil"
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



	httpLog.Trace("New websocket connection")
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

	// Generating UUID
	b := make([]byte, 16)
	rand.Read(b)

	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	// Sending UUID to client
	l.ws.WriteMessage(websocket.TextMessage, []byte("uuid\n" + uuid))

	for {
		_, command, err := l.ws.ReadMessage()
		if err != nil {
			break;
		}

		cmdStr := string(command);
		httpLog.Trace(cmdStr)
		cmdChunk := strings.Split(cmdStr, "\n")
		if len(cmdChunk) != 5 {
			l.ws.WriteMessage(websocket.TextMessage, []byte("Error\nWrong handshake packet"))
		} else {
			realm := cmdChunk[1]
			nonce := cmdChunk[2]
			filter := cmdChunk[3]
			hash := cmdChunk[4]

		}
	}

	l.ws.Close()
}

type HttpServer struct {
	addr		string
	listener 	net.Listener
	stats		interface{}
	dispatch	*Dispatcher
}

func NewHttpServer(addr string, d *Dispatcher) *HttpServer {
	hs := HttpServer{
		addr: addr,
		dispatch: d,
	}

	return &hs
}

func (h *HttpServer) WithStats(stats interface{}) {
	httpLog.Info("Stats available at /stat")
	h.stats = &stats
}

func (h *HttpServer) Start() error {
	var err error

	h.listener, err = net.Listen("tcp", h.addr)
	if err != nil {
		httpLog.Fail(err)
		return err
	}

	httpLog.Infoc("Starting HTTP server on :addr", map[string]interface{}{"addr": h.addr})
	mux := http.NewServeMux()
	mux.HandleFunc("/", wsHandler)
	if h.stats != nil {
		mux.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			json.NewEncoder(w).Encode(h.stats)
		})
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(h.listener)

	return nil
}

func (h *HttpServer) Stop() {
	httpLog.Info("Stopping HTTP server")
	h.listener.Close()
}