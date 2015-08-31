package eos

import (
  "strconv"
)

/// EOS key structure
type EosKey struct {
  Realm     string
  Schema    string
  Tags      []string
  Fqn       string
  HashCode  uint32
}

/// Returns true if EosKey has requested tag
func (k *EosKey) HasTag (tag string) bool {
  for _, v := range k.Tags {
    if v == tag {
      return true
    }
  }

  return false
}

func (k *EosKey) String() string {
  return k.Fqn + " (" + strconv.FormatUint(uint64(k.HashCode), 16) + ")"
}
