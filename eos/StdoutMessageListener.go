package eos

import "fmt"

func PrintMessage(message Message) {
	fmt.Printf("[%s] %s\n", message.Fqn, message.Payload)
}

func NoopMessage(message Message) {}