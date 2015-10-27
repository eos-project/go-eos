package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/eos-project/go-eos/model"
)

// Named type for hash map identities
type HashMapIdentities map[string]string

// Constructs new hash map identities object
func NewHashMapIdentities() HashMapIdentities {
	return map[string]string{}
}

// Adds new auth data
func (ids HashMapIdentities) Add(realm string, secret string) {
	ids[realm] = secret
}

// Authenticates packet
// Returns nil on success, error otherwise
func (ids HashMapIdentities) AuthenticatePacket(p model.Packet) error {
	if _, ok := ids[p.Realm]; !ok {
		return fmt.Errorf("Realm %s not known", p.Realm)
	}

	hasher := sha256.New()
	hasher.Write([]byte(p.Nonce + p.Payload + ids[p.Realm]))
	checksum := hex.EncodeToString(hasher.Sum(nil))

	if checksum != p.Signature {
		return fmt.Errorf("Wrong signature for %s", p.Realm)
	}

	return nil
}
