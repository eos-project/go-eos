package main

import (
	"github.com/eos-project/go-eos/auth"
	"github.com/eos-project/go-eos/net/udp"
	"github.com/eos-project/go-eos/net/ws"
	"github.com/eos-project/go-eos/server"
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
	configReader := cf.ConfigReader{}

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

	err := configReader.ReadJson("eos.json", &mainConfig)
	if err != nil {
		serverLog.Warn("Unable to read configuration file")
		panic(serverLog.Fail(err))
	}

	// Building stats
	stats := server.RuntimeStatistics{}
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
	auths := auth.NewHashMapIdentities()
	for k, v := range mainConfig.Realms {
		auths.Add(k, v)
		serverLog.Debug("Added identity: " + k)
	}

	// Building dispatcher
	dispatcher := server.Dispatcher{
		StatCount: func(value int) {
			stats.ActiveListeners = value
		},
	}

	// UdpConfig
	udpConf := udp.Config{
		Address: mainConfig.Udp.Address,

		Authenticate: auths.AuthenticatePacket,
		Send:         dispatcher.Send,

		BufferSize: mainConfig.Udp.BufferSize,
		PacketSize: mainConfig.Udp.PacketSize,

		StatServe:        stats.UdpPackets.Inc,
		StatErrorAuth:    stats.UdpErrorAuth.Inc,
		StatErrorConnect: stats.UdpErrorConn.Inc,
		StatErrorParse:   stats.UdpErrorParse.Inc,
	}

	// wsConfig
	wsConfig := ws.Config{
		Address:    mainConfig.Http.Address,
		Dispatcher: &dispatcher,
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
		stopper, err := udp.StartListening(udpConf)
		if err != nil {
			serverLog.Fail(err)
			panic(err)
		}
		sigDispatchList = append(sigDispatchList, stopper)
	}

	// Building and starting HTTP server
	if mainConfig.Http.Enabled {
		stopper, err := ws.StartListening(wsConfig)
		if err != nil {
			serverLog.Fail(err)
			panic(err)
		}
		sigDispatchList = append(sigDispatchList, stopper)
	}

	<-done
	log.Dispatcher.Wait()
	os.Exit(1)
}
