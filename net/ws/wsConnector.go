package ws

import (
	"crypto/rand"
	"fmt"
	"github.com/eos-project/go-eos/model"
	"github.com/eos-project/go-eos/server"
	"github.com/gorilla/websocket"
	"strings"
)

type wsConnector struct {
	ws         *websocket.Conn
	d          *server.Dispatcher
	registered bool
	acceptor   server.Listener
}

func StartWsConnector(ws *websocket.Conn, d *server.Dispatcher) {
	// Building and starting new websocket connector
	l := wsConnector{
		ws: ws,
		d:  d,
	}

	ch := make(chan model.Message)

	// Function, that will accept messages and send it to channel for further delivery
	l.acceptor = func(m model.Message) {
		ch <- m
	}

	// Real acceptor
	go func() {
		for m := range ch {
			l.ws.WriteMessage(websocket.TextMessage, []byte("log\n"+m.Key.Path+"\n"+m.Payload))
		}
	}()

	httpLog.Trace("New websocket connection")
	l.accept()
}

func (l *wsConnector) accept() {

	// Generating UUID
	b := make([]byte, 16)
	rand.Read(b)

	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	// Sending UUID to client
	l.ws.WriteMessage(websocket.TextMessage, []byte("uuid\n"+uuid))

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
				l.d.Register(l.acceptor)
			}
		}
	}

	l.ws.Close()

	// Unregistering
	if l.registered {
		l.d.Unregister(l.acceptor)
	}
}
