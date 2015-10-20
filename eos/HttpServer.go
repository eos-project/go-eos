package eos

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/gotterdemarung/go-log/log"
	"net"
	"net/http"
	"strings"
)

var httpLog = log.Context.WithTags("eos", "http")

func (h *HttpServer) wsHandler(w http.ResponseWriter, r *http.Request) {
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
		d:  h.dispatch,
	}

	httpLog.Trace("New websocket connection")
	l.accept()
}

type WebsocketListener struct {
	ch         chan Message
	ws         *websocket.Conn
	d          *Dispatcher
	registered bool
}

func (l *WebsocketListener) accept() {

	// Generating UUID
	b := make([]byte, 16)
	rand.Read(b)

	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	// Sending UUID to client
	l.ws.WriteMessage(websocket.TextMessage, []byte("uuid\n"+uuid))

	// Anonymous function to send messages
	delivery := func(m Message) {
		l.ws.WriteMessage(websocket.TextMessage, []byte("log\n"+m.EosKey.Path+"\n"+m.Payload))
	}

	for {
		_, command, err := l.ws.ReadMessage()
		if err != nil {
			httpLog.Fail(err)
			break
		}

		cmdStr := string(command)
		httpLog.Trace(cmdStr)
		cmdChunk := strings.Split(cmdStr, "\n")
		if len(cmdChunk) != 5 {
			l.ws.WriteMessage(websocket.TextMessage, []byte("error\nWrong handshake packet"))
		} else {
			realm := cmdChunk[1]
			nonce := cmdChunk[2]
			filter := cmdChunk[3]
			hash := cmdChunk[4]

			httpLog.With(
				map[string]interface{}{
					"realm":  realm,
					"nonce":  nonce,
					"filter": filter,
					"hash":   hash,
				},
			).Info("Received websocket auth for realm :realm with nonce :nonce filter :filter signed with :hash")

			// Registering
			l.ws.WriteMessage(websocket.TextMessage, []byte("connected"))
			if !l.registered {
				l.registered = true
				l.d.Register(delivery)
			}
		}
	}

	l.ws.Close()

	// Unregistering
	if l.registered {
		l.d.Unregister(delivery)
	}
}

type HttpServer struct {
	addr     string
	listener net.Listener
	stats    interface{}
	dispatch *Dispatcher
}

func NewHttpServer(addr string, d *Dispatcher) *HttpServer {
	hs := HttpServer{
		addr:     addr,
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

	httpLog.Context["addr"] = h.addr
	httpLog.Info("Starting HTTP server on :addr")
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.wsHandler)
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
