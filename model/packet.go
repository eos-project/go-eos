package model

// Represents incoming packet
type Packet struct {
	Message
	Nonce     string
	Signature string
}
