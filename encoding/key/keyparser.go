package key

import (
	"fmt"
	"github.com/eos-project/go-eos/model"
	"hash/fnv"
	"regexp"
	"sort"
	"strings"
)

var keyParserRegex = regexp.MustCompile("^([a-z0-9\\-_]*)\\+([a-z\\-]*)://(.+)")

// Parses string into EosKey structure
func ParseKey(addr string) (*model.Key, error) {
	matches := keyParserRegex.FindStringSubmatch(addr)

	if len(matches) != 4 {
		return nil, fmt.Errorf("Wrong Eos tracking address format \"%s\"", addr)
	}

	tags := strings.Split(strings.ToLower(matches[3]), ":")
	sort.Strings(tags)

	key := model.Key{
		Realm:  matches[1],
		Schema: matches[2],
		Tags:   tags,
	}
	key.Path = key.Schema + "://" + strings.Join(key.Tags, ":")
	key.Fqn = key.Realm + "+" + key.Path

	h := fnv.New32a()
	h.Write([]byte(key.Fqn))
	key.HashCode = h.Sum32()

	return &key, nil
}
