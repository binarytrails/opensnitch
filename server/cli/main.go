package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/storage"
	"github.com/gustavo-iniguez-goya/opensnitch/server/cli/views"
)

type viewType string

var (
	// addrs:port, :port -> anyAddrs:port
	daemonMode           = false
	serverPort           = ":50051"
	serverProto          = "tcp"
	sigChan              = (chan os.Signal)(nil)
	exitChan             = (chan bool)(nil)
	showInteractiveShell = false
	viewsConfig          *views.Config
	showStatus           = false
	dbType               = storage.Sqlite
	dbDSN                = "file::memory:?cache=shared"
	dbDebug              = false
)

func setupSignals() {
	sigChan = make(chan os.Signal)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt)
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

	viewsConfig = views.New()

	// TODO: move out to a .json configuration file
	flag.IntVar(&dbType, "db-type", storage.Sqlite, "Db type to save the events: 0 (sqlite), 1 (postgresql), 2 (mysql)")
	flag.StringVar(&dbDSN, "db-dsn", dbDSN, "db connection string. Example: postgres: \"host=localhost user=postgres password=postgres dbname=gorm port=5432 sslmode=disable\", sqlite: /tmp/opensnitch.db or \""+dbDSN+"\", mysql: \"\"")
	flag.BoolVar(&dbDebug, "db-debug", false, "Debug DB slow queries")

	flag.StringVar(&serverProto, "socket-type", "tcp", "Protocol for incoming nodes (tcp, udp, unix)")
	flag.StringVar(&serverPort, "socket-port", ":50051", "Listening port for incoming nodes (127.0.0.1:50051, :50051, /tmp/osui.sock")
	flag.StringVar(&viewsConfig.View, "show-stats", "", "View connections statistics, possible values: general, nodes, hosts, procs, addrs, ports, users, rules, nodes")
	flag.StringVar(&viewsConfig.Delimiter, "stats-delimiter", "", "Delimiter to separate statistics fields when print style is 'plain'")
	flag.IntVar(&viewsConfig.Limit, "stats-limit", -1, "Limit statistics")
	flag.StringVar(&viewsConfig.Style, "stats-style", views.ViewStylePretty, "Lists style: pretty, plain")
	flag.StringVar(&viewsConfig.Filter, "stats-filter", "", "Filter statistics. For example: firefox")
	flag.BoolVar(&viewsConfig.Loop, "live", true, "Live statistics. If false, only last statistics of nodes will be printed")
	flag.BoolVar(&daemonMode, "D", false, "Work as daemon, no UI")
	flag.BoolVar(&showStatus, "show-status", false, "Show daemon status and exit")
	// TODO: stats-fields: time,proc,dstIp,dstPort ...
}

func usage() {
	fmt.Printf("\n Options:\n")
	flag.PrintDefaults()
	fmt.Println("\n Usage:")
	fmt.Printf(" ./opensnitch-cli -show-stats general -socket-type tcp -socket-port :50052\n")
	fmt.Printf(" ./opensnitch-cli -show-stats general -socket-type unix -socket-port /tmp/osui.sock\n")
	fmt.Printf(" ./opensnitch-cli -show-stats general -db-type 0 -db-dsn /tmp/opensnitch.db\n")

	fmt.Println("\n Dump connection events with the given delimiter:")
	fmt.Printf(" ./opensnitch-cli -show-stats general -socket-type tcp -socket-port :50052 -stats-style plain -stats-delimiter , -live=false\n")

	fmt.Println("\n Work as daemon:")
	fmt.Printf(" ./opensnitch-cli -D\n")

	fmt.Println("\n Print only nodes statistics:")
	fmt.Printf(" ./op-cli -socket-type tcp -socket-port :50051 -db-type 0 -db-dsn /tmp/opensnitch.db\n.")
	println()
}

func main() {
	flag.Parse()
	setupSignals()

	if flag.NFlag() == 0 {
		usage()
		return
	}

	st := storage.NewStorage(dbDSN, storage.DbType(dbType), dbDebug)
	apiClient := api.NewClient(serverProto, serverPort, daemonMode, st)
	if daemonMode {
		api.StartServer(apiClient, serverProto, serverPort)
		return
	}

	views.Init(apiClient, viewsConfig)
	//apiClient.WaitForNodes()

	if showStatus {
		views.PrintStatus()
		return
	}
	if viewsConfig.View != "" {
		views.Show()
	}

	views.RestoreTTY()
}
