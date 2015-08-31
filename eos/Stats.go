package eos

import (
	"sync/atomic"
)

type StatCounter struct {
	Value int64
}

func (s *StatCounter) Inc() {
	atomic.AddInt64(&s.Value, 1)
}

type RuntimeStatistics struct {
	UdpPackets 		StatCounter
	UdpErrorConn	StatCounter
	UdpErrorParse	StatCounter
	UdpErrorAuth	StatCounter
}