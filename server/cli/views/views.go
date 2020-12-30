package views

import (
	"fmt"
	"sync"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
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
	apiClient *api.Client
}

var (
	lock          sync.RWMutex
	config        *Config
	ui            *UI
	screen        *Screen
	bView         *BaseView
	vNodes        *ViewNodes
	vEvents       *ViewEvents
	vEventsByType *ViewEventsByType
	vRules        *ViewRules

	keyPressedChan chan *menus.KeyEvent

	viewNames = [...]string{
		ViewNameGeneral,
		ViewNameHosts,
		ViewNameProcs,
		ViewNameAddrs,
		ViewNamePorts,
		ViewNameUsers,
		ViewNameRules,
		ViewNameNodes,
	}
	viewList = map[string]int8{
		ViewNameGeneral: 0,
		ViewNameHosts:   1,
		ViewNameProcs:   2,
		ViewNameAddrs:   3,
		ViewNamePorts:   4,
		ViewNameUsers:   5,
		ViewNameRules:   6,
		ViewNameNodes:   7,
	}
	totalViews = int8(8)
)

const (
	appName    = "OpenSnitch"
	appVersion = "0.0.1"

	ViewNameGeneral = "general"
	ViewNameHosts   = "hosts"
	ViewNameProcs   = "procs"
	ViewNameAddrs   = "addrs"
	ViewNamePorts   = "ports"
	ViewNameUsers   = "users"
	ViewNameRules   = "rules"
	ViewNameNodes   = "nodes"

	ViewStylePlain  = "plain"
	ViewStylePretty = "pretty"

	defaultRulesTimeout = api.Rule15s
)

// New returns an empty Config struct
func New() *Config {
	return &Config{}
}

// Init configures the views.
func Init(apiClient *api.Client, conf *Config) {
	config = conf
	config.apiClient = apiClient

	bView = NewBaseView()
	ui = NewUI(apiClient, conf)
	screen = NewScreen(ui, apiClient)

	vNodes = NewViewNodes(screen, bView)
	vEvents = NewViewEvents(screen, bView)
	vEventsByType = NewViewByType(screen, bView)
	vRules = NewViewRules(screen, bView)

	keyPressedChan = menus.Interactive()

	//go handleNewRules()
}

// Show displays the given statistics.
func Show() {
	switch config.View {
	case ViewNameGeneral:
		vEvents.Print()
	case ViewNameHosts, ViewNameProcs, ViewNameAddrs, ViewNamePorts, ViewNameUsers:
		vEventsByType.Print(viewList[config.View], config.View)
	case ViewNameNodes:
		vNodes.Print()
	case ViewNameRules:
		vRules.Print()
	default:
		log.Info("unknown view:", config.View)
	}
}

func PrintStatus() {
	screen.PrintStatus()
}

func exit() {
	ui.cleanLine()
	menus.Exit()

	lock.Lock()
	defer lock.Unlock()
	ui.stopStats = true
}

// just after start the stats are not ready to be consumed, because there're no
// nodes connected, so we need to wait for new nodes.

func setFilter() {
	ui.cleanLine()
	if config.Filter != "" {
		fmt.Printf("Current filter: %s\n", config.Filter)
	}
	screen.printPrompt("filter")
	config.Filter = menus.ReadLine()
}

func unsetFilter() {
	config.Filter = ""
}

func menuGeneral(key *menus.KeyEvent) {
	switch key.Key {
	case menus.NEXTVIEWARROW:
		nextView()
		return
	case menus.PREVVIEWARROW:
		prevView()
		return
	case menus.SORTASCENDING:
		vEvents.sortAscending()
		return
	case menus.SORTDESCENDING:
		vEvents.sortDescending()
		return
	}
	switch key.Char {
	case menus.PAUSE:
		ui.pauseStats = true
	case menus.CONTINUE, menus.RUN:
		ui.pauseStats = false
	case menus.QUIT:
		ui.pauseStats = true
		exit()
		return
	case menus.HELP:
		ui.pauseStats = true
		screen.printHelp()
	case menus.NEXTVIEW:
		nextView()
	case menus.PREVVIEW:
		prevView()
	case menus.FILTER:
		setFilter()
	case menus.DISABLEFILTER:
		unsetFilter()
	case menus.ACTIONS:
		ui.pauseStats = true
		screen.printActionsMenu()
	case menus.STOPFIREWALL:
		nodes.StopFirewall()
		ui.pauseStats = false
	case menus.STARTFIREWALL:
		nodes.StartFirewall()
		ui.pauseStats = false
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
	case key, ok := <-menus.KeyPressedChan:
		if !ok || key == nil || key.Char == "" {
			return
		}
		menuGeneral(key)
	default:
	}
}

// start listening for new rules and ask the user to allow or deny them.
func handleNewRules() {
	for {
		if ui.getStopStats() {
			return
		}
		select {
		case con := <-ui.aClient.WaitForRules():
			vRules.askRule(con)
		}
	}
}
