package views

import (
	"fmt"
	"sync"
	"time"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"github.com/gustavo-iniguez-goya/opensnitch/server/cli/menus"
)

type Config struct {
	View      string
	Delimiter string
	Loop      bool
	Style     string
	Limit     int
	Filter    string
	//Fields    []string
	apiClient *api.Client
}

var (
	lock   sync.RWMutex
	config Config

	pauseStats = false
	stopStats  = false

	ttyRows = 80
	ttyCols = 80

	keyPressedChan chan *menus.KeyEvent

	viewNames = [...]string{
		ViewGeneral,
		ViewHosts,
		ViewProcs,
		ViewAddrs,
		ViewPorts,
		ViewUsers,
		ViewRules,
		ViewNodes,
	}
	viewList = map[string]int{
		ViewGeneral: 0,
		ViewHosts:   1,
		ViewProcs:   2,
		ViewAddrs:   3,
		ViewPorts:   4,
		ViewUsers:   5,
		ViewRules:   6,
		ViewNodes:   7,
	}
	totalViews = 7
)

const (
	appName     = "OpenSnitch"
	ViewGeneral = "general"
	ViewHosts   = "hosts"
	ViewProcs   = "procs"
	ViewAddrs   = "addrs"
	ViewPorts   = "ports"
	ViewUsers   = "users"
	ViewRules   = "rules"
	ViewNodes   = "nodes"

	ViewStylePlain  = "plain"
	ViewStylePretty = "pretty"
	// The extra %s corresponds to the delimiter, which for this view is always ""
	printFormatPretty = "%v %s[ %-22s]%s [%-4s]%s [%-5d]%s - %-5v%s:%-15v%s -> %16v%s:%-5v%s - %s%s\n"
	printFormatPlain  = "%v%s%s%s%s%s%d%s%v%s%v%s%v%s%v%s%s%s\n"

	defaultRulesTimeout = api.Rule15s
)

// Init configures the views.
func Init(apiClient *api.Client, conf Config) {
	config = conf
	config.apiClient = apiClient
	keyPressedChan = menus.Interactive()

	go handleNewRules()
	getTTYSize()
}

// Show displays the given statistics.
func Show() {
	switch config.View {
	case ViewGeneral:
		GeneralStats()
	case ViewHosts, ViewProcs, ViewAddrs, ViewPorts, ViewUsers:
		StatsByType(config.View)
	case ViewNodes:
		NodesList()
	case ViewRules:
		RulesList()
	default:
		log.Info("unknown view:", config.View)
	}
}

func exit() {
	cleanLine()
	menus.Exit()

	lock.Lock()
	defer lock.Unlock()
	stopStats = true
}

func getPauseStats() bool {
	lock.RLock()
	defer lock.RUnlock()

	return pauseStats
}

func getStopStats() bool {
	lock.RLock()
	defer lock.RUnlock()

	return stopStats || !config.Loop
}

// just after start the stats are not ready to be consumed, because there're no
// nodes connected, so we need to wait for new nodes.
func waitForStats() {
	tries := 30
	for config.apiClient.GetLastStats() == nil && tries > 0 {
		time.Sleep(1 * time.Second)
		tries--
		log.Raw(log.Wrap("No stats yet, waiting ", log.GREEN)+" %d\r", tries)
	}
	cleanLine()
}

func setFilter() {
	cleanLine()
	if config.Filter != "" {
		fmt.Printf("Current filter: %s\n", config.Filter)
	}
	fmt.Printf(log.BG_GREEN + log.FG_WHITE + "filter" + log.Bold(">") + log.RESET + " ")
	config.Filter = menus.ReadLine()
}

func unsetFilter() {
	config.Filter = ""
}

func menuGeneral(cmd string) {
	switch cmd {
	case menus.PAUSE:
		pauseStats = true
	case menus.CONTINUE, menus.RUN:
		pauseStats = false
	case menus.QUIT:
		pauseStats = true
		exit()
		return
	case menus.HELP:
		pauseStats = true
		printHelp()
	case menus.NEXTVIEW:
		nextView()
	case menus.PREVVIEW:
		prevView()
	case menus.FILTER:
		setFilter()
	case menus.DISABLEFILTER:
		unsetFilter()
	case menus.ACTIONS:
		pauseStats = true
		printActionsMenu()
	case menus.STOPFIREWALL:
		nodes.SendNotifications(
			&protocol.Notification{
				Id:   uint64(time.Now().UnixNano()),
				Type: protocol.Action_UNLOAD_FIREWALL,
			})
		pauseStats = false
	case menus.STARTFIREWALL:
		nodes.SendNotifications(
			&protocol.Notification{
				Id:   uint64(time.Now().UnixNano()),
				Type: protocol.Action_LOAD_FIREWALL,
			})
		pauseStats = false
	// case menus.CHANGECONFIG:
	// case menus.DELETERULE
	// TODO
	// case menus.ShowViewsMenu (: -> :ls, :hosts, :ports, :procs, :users, ...)
	//  - case menus.ViewByHost
	//  - case menus.ViewByPort
	//  - case menus.ViewConnectionDetails -> select row -> show details
	//  - case menus.SortReverse (o)
	default:
	}
}

func readLiveMenu() {
	select {
	case cmd, ok := <-menus.KeyPressedChan:
		if !ok || cmd.Char == "" {
			return
		}
		menuGeneral(cmd.Char)
	default:
	}
}

// start listening for new rules and ask the user to allow or deny them.
func handleNewRules() {
	for {
		if getStopStats() {
			return
		}
		select {
		case con := <-config.apiClient.WaitForRules():
			askRule(con)
		}
	}
}
