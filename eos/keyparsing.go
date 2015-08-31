package eos

import (
	"fmt"
	"strings"
	"regexp"
	"sort"
	"hash/fnv"
)

var keyParserRegex = regexp.MustCompile("^([a-z0-9\\-_]*)\\+([a-z\\-]*)://(.+)")

func ParseKey(addr string) (*EosKey, error) {
	matches := keyParserRegex.FindStringSubmatch(addr)

	if len(matches) != 4 {
		return nil, fmt.Errorf("Wrong Eos tracking address format \"%s\"", addr)
	}

	tags := strings.Split(strings.ToLower(matches[3]), ":")
	sort.Strings(tags)

	key := EosKey{};
	key.Realm = matches[1]
	key.Schema = matches[2]
	key.Tags = tags
	key.Fqn = key.Realm + "+" + key.Schema + "://" + strings.Join(key.Tags, ":")

	h := fnv.New32a()
	h.Write([]byte(key.Fqn))
	key.HashCode = h.Sum32();

	return &key, nil;
}