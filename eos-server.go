package main

import (
	"github.com/eos-project/go-eos/eos"
	cf "github.com/gotterdemarung/go-configfile"
	"github.com/gotterdemarung/go-log/log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	serverLog := log.Context.WithTags("eos")
	log.Autoconfig("eos")

	serverLog.Context["pid"] = os.Getpid()
	serverLog.Info("Starting EOS server with pid :pid")

	// Loading configuration file
	confFile, err := cf.NewConfigFile("eos.json", true)
	if err != nil {
		panic(serverLog.Fail(err))
	}
	serverLog.Context["full"] = confFile.FullPath
	serverLog.Info("Using config at :full")

	var mainConfig struct {
		Timer  int
		Realms map[string]string
		Udp    struct {
			Enabled    bool
			Address    string
			PacketSize int
			BufferSize int
		}
		Http struct {
			Enabled bool
			Stats   bool
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
			rps := float32(stats.UdpPackets.Value-last) / float32(mainConfig.Timer)
			last = stats.UdpPackets.Value

			serverLog.Context["gor"] = runtime.NumGoroutine()
			serverLog.Context["rps"] = rps
			serverLog.Context["us"] = stats.UdpPackets.Value
			serverLog.Context["uec"] = stats.UdpErrorConn.Value
			serverLog.Context["uep"] = stats.UdpErrorParse.Value
			serverLog.Context["uea"] = stats.UdpErrorAuth.Value

			serverLog.Debug("Goroutines :gor, Udp served :us (:rps RPS) - Conn err: :uec - Parse err: :uep - Auth err: :uea")
		}
	}()

	// Building authenticator
	auth := eos.NewHashMapIdentities()
	for k, v := range mainConfig.Realms {
		auth.Add(k, v)
		serverLog.Debug("Added identity: " + k)
	}

	// Building dispatcher
	dispatcher := eos.Dispatcher{
		StatCount: func(value int) {
			stats.ActiveListeners = value
		},
	}

	// UdpConfig
	udpConf := eos.UdpServerConfiguration{
		Address: mainConfig.Udp.Address,

		ParseKey:     eos.ParseKey,
		Authenticate: auth.AuthenticatePacket,
		Send:         dispatcher.Send,

		BufferSize: mainConfig.Udp.BufferSize,
		PacketSize: mainConfig.Udp.PacketSize,

		StatServe:        stats.UdpPackets.Inc,
		StatErrorAuth:    stats.UdpErrorAuth.Inc,
		StatErrorConnect: stats.UdpErrorConn.Inc,
		StatErrorParse:   stats.UdpErrorParse.Inc,
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
		serverLog.Context["sig"] = sig.String()
		serverLog.Context["dc"] = len(sigDispatchList)
		serverLog.Warn("Received signal :sig, shutting down gracefully. Dispatch list contains :dc funcs")
		for _, f := range sigDispatchList {
			serverLog.Info("Running signal dispatcher")
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

	<-done
	log.Dispatcher.Wait()
	os.Exit(1)
}
