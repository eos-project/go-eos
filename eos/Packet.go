package eos

type Packet struct {
	Message
	Nonce 		string
	Signature	string
}