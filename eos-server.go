package main

import (
	"time"
	"os"
	"os/signal"
	"syscall"
	"runtime"
	"github.com/gotterdemarung/go-log/log"
	cf "github.com/gotterdemarung/go-configfile"
  	"github.com/eos-project/go-eos/eos"
)

func main() {
	serverLog := log.Context.WithTags("eos")
	log.Dispatcher.FromCli()

	serverLog.Infoc("Starting EOS server with pid :pid", map[string]interface{}{"pid": os.Getpid()});

	// Loading configuration file
	confFile, err := cf.NewConfigFile("eos.json", true)
	if err != nil {
		panic(serverLog.Fail(err))
	}
	serverLog.Infoc("Using config at :full", map[string]interface{}{"full": confFile.FullPath})

	var mainConfig struct {
		Timer int
		Realms map[string]string
		Udp struct {
			Enabled bool
			Address string
			PacketSize int
			BufferSize int
		}
		Http struct {
			Enabled bool
			Stats bool
			Address string
		}
	}

	err = confFile.DecodeJson(&mainConfig)
	if err != nil {
		serverLog.Warn("Unable to read configuration file")
		panic(serverLog.Fail(err))
	}

	// Building stats
	stats := eos.RuntimeStatistics{}
	statsTicker := time.NewTicker(time.Duration(mainConfig.Timer) * time.Second)
	go func() {
		var last int64
		last = 0

		for _ = range statsTicker.C {
			rps := float32(stats.UdpPackets.Value - last) / float32(mainConfig.Timer)
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
	for k,v := range mainConfig.Realms {
		auth.Add(k, v)
		serverLog.Debugc("Added identity :name", map[string]interface{}{"name": k})
	}

	// Building dispatcher
	dispatcher := eos.Dispatcher{
		StatCount: func(value int) {
			stats.ActiveListeners = value
		},
	}

	// UdpConfig
	udpConf := eos.UdpServerConfiguration{
		Address:		mainConfig.Udp.Address,

		ParseKey:		eos.ParseKey,
		Authenticate:	auth.AuthenticatePacket,
		Send:			dispatcher.Send,

		BufferSize:			mainConfig.Udp.BufferSize,
		PacketSize:			mainConfig.Udp.PacketSize,

		StatServe:			stats.UdpPackets.Inc,
		StatErrorAuth:  	stats.UdpErrorAuth.Inc,
		StatErrorConnect:	stats.UdpErrorConn.Inc,
		StatErrorParse:		stats.UdpErrorParse.Inc,
	}

	// Signals dispatchering
	sigDispatchList := []func(){
		statsTicker.Stop,
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	done := make(chan bool)
	go func() {
		sig := <-c
		serverLog.Warnc(
			"Received signal :sig, shutting down gracefully. Dispatch list contains :count funcs",
			map[string]interface{}{"sig": sig.String(), "count": len(sigDispatchList)},
		)
		for i, f := range sigDispatchList {
			serverLog.Infoc("Running signal dispatcher # :ii", map[string]interface{}{"ii": i})
			f()
		}
		serverLog.Info("Done with dispatchers")

		done <- true
	}()

	// Building and starting UDP listener
	if mainConfig.Udp.Enabled {
		udp, err := eos.NewUdpServer(udpConf)
		if err != nil {
			serverLog.Fail(err)
			panic(err)
		}
		err = udp.Start()
		if err != nil {
			serverLog.Fail(err)
			panic(err)
		}
		sigDispatchList = append(sigDispatchList, udp.Stop)
	}

	// Building and starting HTTP server
	if mainConfig.Http.Enabled {
		hs := eos.NewHttpServer(mainConfig.Http.Address, &dispatcher)
		if mainConfig.Http.Stats {
			hs.WithStats(&stats)
		}
		err := hs.Start()
		if err != nil {
			serverLog.Fail(err)
			panic(err)
		}
		sigDispatchList = append(sigDispatchList, hs.Stop)
	}

	// Add demo listener
	dispatcher.Register(eos.NoopMessage)

	<- done
	os.Exit(1)
}
