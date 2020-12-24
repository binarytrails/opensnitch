package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/cli/views"
)

type viewType string

var (
	// addrs:port, :port -> anyAddrs:port
	serverPort           = ":50051"
	serverProto          = "tcp"
	sigChan              = (chan os.Signal)(nil)
	exitChan             = (chan bool)(nil)
	showInteractiveShell = false
	viewsConfig          views.Config
	showStatus           = false
)

func setupSignals() {
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		sig := <-sigChan
		log.Raw("\n")
		log.Important("Got signal: %v", sig)
		//		exitChan <- true
		views.RestoreTTY()
		os.Exit(0)
	}()
}

func init() {
	flag.StringVar(&serverProto, "socket-type", "tcp", "Protocol for incoming nodes (tcp, udp, unix)")
	flag.StringVar(&serverPort, "socket-port", ":50051", "Listening port for incoming nodes (127.0.0.1:50051, :50051, /tmp/osui.sock")
	flag.StringVar(&viewsConfig.View, "show-stats", "", "View connections statistics, possible values: general, nodes, hosts, procs, addrs, ports, users, rules, nodes")
	flag.StringVar(&viewsConfig.Delimiter, "stats-delimiter", "", "Delimiter to separate statistics fields when print style is 'plain'")
	flag.IntVar(&viewsConfig.Limit, "stats-limit", -1, "Limit statistics")
	flag.StringVar(&viewsConfig.Style, "stats-style", views.ViewStylePretty, "Lists style: pretty, plain")
	flag.StringVar(&viewsConfig.Filter, "stats-filter", "", "Filter statistics. For example: firefox")
	flag.BoolVar(&viewsConfig.Loop, "live", true, "Live statistics. If false, only last statistics of nodes will be printed")
	flag.BoolVar(&showStatus, "show-status", false, "Show daemon status and exit")
	// TODO: stats-fields: time,proc,dstIp,dstPort ...
}

func main() {
	flag.Parse()
	setupSignals()

	if flag.NFlag() == 0 {
		fmt.Printf("\n Options:\n")
		flag.PrintDefaults()
		fmt.Println(" Usage:")
		fmt.Printf("\n ./opensnitch-cli -show-stats general -socket-type tcp -socket-port :50052\n")
		fmt.Printf(" ./opensnitch-cli -show-stats general -socket-type tcp -socket-port :50052 -stats-style plain -stats-delimiter , -live=false\n")
		println()
		return
	}

	apiClient := api.NewClient(serverProto, serverPort)
	// TODO: work as a daemon, no interactive views, no data printed to the terminal
	views.Init(apiClient, viewsConfig)
	// Start the server and wait for incoming nodes
	log.Info("Waiting for nodes...")
	apiClient.WaitForNodes()
	log.Info("Ready")

	// With at least 1 node, display the stats
	if showStatus {
		views.PrintStatus()
	} else if viewsConfig.View != "" {
		views.Show()
	}

	views.RestoreTTY()
}
