package model

import (
	"strconv"
)

// EOS key structure
type Key struct {
	Realm    string
	Schema   string
	Tags     []string
	Fqn      string
	Path     string
	HashCode uint32
}

// Returns true if EosKey has requested tag
func (k *Key) HasTag(tag string) bool {
	for _, v := range k.Tags {
		if v == tag {
			return true
		}
	}

	return false
}

// Returns string representation of eos key
func (k Key) String() string {
	return k.Fqn + " (" + strconv.FormatUint(uint64(k.HashCode), 16) + ")"
}
