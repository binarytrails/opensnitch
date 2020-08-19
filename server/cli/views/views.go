package views

import (
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/cli/menus"
)

type Config struct {
	View      string
	Delimiter string
	Loop      bool
	Style     string
	Limit     int
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

	keyPressedChan = make(chan string, 1)

	viewNames = [...]string{
		ViewGeneral,
		ViewHosts,
		ViewProcs,
		ViewAddrs,
		ViewPorts,
		ViewUsers,
	}
	viewList = map[string]int{
		ViewGeneral: 0,
		ViewHosts:   1,
		ViewProcs:   2,
		ViewAddrs:   3,
		ViewPorts:   4,
		ViewUsers:   5,
	}
	totalViews = 5
)

const (
	appName     = "OpenSnitch"
	ViewGeneral = "general"
	ViewNodes   = "nodes"
	ViewHosts   = "hosts"
	ViewProcs   = "procs"
	ViewAddrs   = "addrs"
	ViewPorts   = "ports"
	ViewUsers   = "users"

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
	case ViewNodes, ViewHosts, ViewProcs, ViewAddrs, ViewPorts, ViewUsers:
		StatsByType(config.View)
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

	return stopStats
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

func readLiveMenu() {
	select {
	case cmd, ok := <-menus.KeyPressedChan:
		if !ok || cmd == "" {
			return
		}
		switch cmd {
		case menus.PAUSE:
			pauseStats = true
		case menus.CONTINUE, menus.RUN:
			pauseStats = false
		case menus.QUIT:
			exit()
			return
		case menus.HELP:
			pauseStats = true
			printHelp()
		case menus.NEXTVIEW:
			nextView()
		case menus.PREVVIEW:
			prevView()
			// TODO
			// case menus.ShowViewsMenu (: -> :ls, :hosts, :ports, :procs, :users, ...)
			//  - case menus.ViewByHost
			//  - case menus.ViewByPort
			//  - case menus.ViewConnectionDetails -> select row -> show details
			//  - case menus.SortReverse (o)
		default:
		}
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

func askRule(con *protocol.Connection) {
	lock.Lock()
	defer lock.Unlock()
	pauseStats = true

	procName := path.Base(con.ProcessPath)
	// TODO: uglify rule name
	ruleName := fmt.Sprint(procName, "-", con.Protocol, "-sport", con.SrcPort, "-dport", con.DstPort)
	alertTitle := log.Bold(log.Red(fmt.Sprintf("**** %s is trying to establish a connection ****", procName)))
	alertBody := log.Blue(fmt.Sprint(procName, ": ", con.SrcPort, ":", con.SrcIp, " -> ", con.DstIp, ":", con.DstPort))
	alertButtons := log.Bold(fmt.Sprint(log.Green("✔ Allow(1)"), ", ", log.Bold(log.Red("✘ Deny(2)")), ", ", log.Bold("Connection details(3)")))

	// TODO: add more options: filter by fields, regexp rule, etc.
	if con.ProcessPath == "" {
		alertTitle = log.Bold(log.Red("  **** New outgoing connection ****  "))
	}

	questionBox(alertTitle, alertBody, alertButtons)

	timeout, _ := time.ParseDuration(defaultRulesTimeout)
	time.AfterFunc(timeout, func() {
		log.Important("Timeout, default action applied\n")
		setBlinkingLabel(RULES)
		menus.KeyPressedChan <- menus.NotAnswered
	})
	cleanLine()
	switch key := <-menus.KeyPressedChan; key {
	case menus.Allow:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
		resetBlinkingLabel(RULES)
	case menus.Deny:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Red("✘ Connection DENIED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionDeny)
		resetBlinkingLabel(RULES)
		// TODO
		// case menus.ShowRuleOptions
	case menus.NotAnswered:
		// TODO: configure default action
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
	}
	time.Sleep(1 * time.Second)

	pauseStats = false

	op := &protocol.Operator{
		Type:    api.RuleSimple,
		Operand: api.FilterByPath,
		Data:    con.ProcessPath,
	}
	config.apiClient.AddNewRule(&protocol.Rule{
		Name:     ruleName,
		Enabled:  true,
		Action:   api.ActionAllow,
		Duration: api.Rule30s,
		Operator: op,
	})

}
