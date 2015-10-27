package native

import (
	"bytes"
	"fmt"
	"github.com/eos-project/go-eos/encoding/key"
	"github.com/eos-project/go-eos/model"
	"strings"
)

var nl []byte = []byte("\n")

// Encodes packet into byte slice
func MarshalPacket(v model.Packet) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(v.Nonce)
	buf.Write(nl)
	buf.WriteString(v.Signature)
	buf.Write(nl)
	buf.WriteString(v.Key.Fqn)
	buf.Write(nl)
	buf.WriteString(v.Payload)
	return buf.Bytes()
}

// Decodes bytes into known packet
func UnmarshalPacket(data []byte) (*model.Packet, error) {
	if data == nil || len(data) == 0 {
		return nil, fmt.Errorf("Incoming bytes empty")
	}

	chunks := strings.Split(string(data), "\n")
	if len(chunks) < 4 {
		return nil, fmt.Errorf("Wrong chunks size")
	}

	nonce := chunks[0]
	signature := chunks[1]
	keyString := chunks[2]

	if i := strings.Index(signature, "+"); i != -1 {
		// Realm in signature
		keyString = signature[0:i+1] + keyString
		signature = signature[i+1:]
	}

	// Resolving key
	key, err := key.ParseKey(keyString)
	if err != nil {
		return nil, err
	}

	msg := model.Message{
		*key,
		strings.TrimSpace(strings.Join(chunks[3:], "\n")),
	}
	return &model.Packet{
		msg,
		nonce,
		signature,
	}, nil
}
