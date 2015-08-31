package main

import (
  	"fmt"
	"time"
	"runtime"
	"github.com/gotterdemarung/go-log/log"
  	"github.com/eos-project/go-eos/eos"
)

func main() {
	serverLog := log.Context.WithTags("eos")
	log.Dispatcher.FromCli()

	serverLog.Info("Starting EOS server");

	// Building stats
	var timerSec time.Duration
	timerSec = 5
	stats := eos.RuntimeStatistics{}
	go func() {
		var last int64
		last = 0

		for _ = range time.Tick(timerSec * time.Second) {
			rps := float32(stats.UdpPackets.Value - last) / float32(timerSec)
			last = stats.UdpPackets.Value

			serverLog.Debugc(
				"Goroutines :gor, Udp served :us (:rps RPS) - :uec - :uep - :uea",
				map[string]interface{}{
					"gor":	runtime.NumGoroutine(),
					"rps":  rps,
					"us": 	stats.UdpPackets.Value,
					"uec":	stats.UdpErrorConn.Value,
					"uep":	stats.UdpErrorParse.Value,
					"uea":	stats.UdpErrorAuth.Value,
				},
			)
		}
	}()

  	// Building authenticator
	auth := eos.NewHashMapIdentities()
	auth.Add("guest", "guest")
	auth.Add("xxx", "yyy")
	serverLog.Debugc("Added guest identity :name", map[string]interface{}{"name": "guest"})

	// Building dispatcher
	dispatcher := eos.Dispatcher{}

	// UdpConfig
	udpConf := eos.UdpServerConfiguration{
		Address:		":8087",

		ParseKey:		eos.ParseKey,
		Authenticate:	auth.AuthenticatePacket,
		Send:			dispatcher.Send,

		StatServe:			stats.UdpPackets.Inc,
		StatErrorAuth:  	stats.UdpErrorAuth.Inc,
		StatErrorConnect:	stats.UdpErrorConn.Inc,
		StatErrorParse:		stats.UdpErrorParse.Inc,
	}

	// Building HTTP server
//	hs := eos.NewHttpServer(":8090")

	// Building Udp listener
	udp, err := eos.NewUdpServer(udpConf)
	if err != nil {
		panic(err)
	}

	// Starting servers
	udp.Start()
//	hs.Start()

	// Add demo listener
//	dispatcher.Register(eos.PrintMessage)
	dispatcher.Register(eos.NoopMessage)
//	dispatcher.Register(eos.PrintMessage)

	var input string
	fmt.Scanln(&input)
	fmt.Println("done")
}
