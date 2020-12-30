package views

import (
	"fmt"
	"time"

	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/storage"
)

var (
	actionDeny  = log.Bold(log.Red(api.ActionDeny))
	actionAllow = log.Bold(log.Green(api.ActionAllow))
	eventAction = actionDeny
	eventTime   = ""
)

// Screen holds the functionality for printing data to the terminal.
type Screen struct {
	*UI
	aClient *api.Client

	actionDeny  string
	actionAllow string
	eventAction string
	printFormat string
	eventTime   string

	printFormatPretty string
	printFormatPlain  string

	totalViews uint8
}

// NewScreen initializes a new Screen struct with default values.
func NewScreen(sui *UI, aClient *api.Client) *Screen {
	scr := &Screen{
		ui,
		aClient,
		log.Bold(log.Red(api.ActionDeny)),
		log.Bold(log.Green(api.ActionAllow)),
		actionDeny,
		"",
		"",
		"%v %s [%-20s]%s[ %-22s]%s [%-4s]%s [%-5d]%s - %-5v%s:%-15v%s -> %16v%s:%-5v%s - %s%s\n",
		// The extra %s corresponds to the delimiter, which for this view is always ""
		"%v%s%s%s%s%s%s%s%d%s%v%s%v%s%v%s%v%s%s%s\n",
		7,
	}

	scr.printFormat = scr.printFormatPretty
	return scr
}

func (s *Screen) printPrompt(text string) {
	s.cleanLine()
	fmt.Printf(log.BG_GREEN + log.FG_WHITE + text + log.Bold(">") + log.RESET + " ")
}

func (s *Screen) printVerticalPadding(what int) {
	if what < int(s.ttyRows) {
		for row := 0; row < int(s.ttyRows)-what-2; row++ {
			fmt.Printf("\n")
		}
	}
}

// PrintStatus prints an overview of the node statistics, and exit.
func (s *Screen) PrintStatus() {
	s.waitForStats()

	dbNodes := s.aClient.GetNodeStats()

	appTitle := fmt.Sprint(UNDERLINE, appName+" - "+appVersion+log.RESET)

	fmt.Printf("\n\t%s\n\n", log.Bold(appTitle))

	for _, node := range *dbNodes {
		uptime, _ := time.ParseDuration(fmt.Sprint(node.Statistics.Uptime, "s"))
		fmt.Printf(
			"\tAddress:\t%s\n"+
				"\tHost:\t\t%s\n"+
				"\tVersion:\t%s\n"+
				"\tRules:\t\t%d\n"+
				"\tUptime:\t\t%s\n"+
				"\tDnsResponses:\t%d\n"+
				"\tConnections:\t%d\n"+
				"\tIgnored:\t%d\n"+
				"\tAccepted:\t%d\n"+
				"\tDropped:\t%d\n"+
				"\tRules hits:\t%d\n"+
				"\tRules missed:\t%d\n\n",
			node.Statistics.Node,
			node.Name,
			node.Statistics.DaemonVersion,
			node.Statistics.Rules,
			uptime,
			node.Statistics.DNSResponses,
			node.Statistics.Connections,
			node.Statistics.Ignored,
			node.Statistics.Accepted,
			node.Statistics.Dropped,
			node.Statistics.RuleHits,
			node.Statistics.RuleMisses,
		)
	}
}

func (s *Screen) printHelp() {
	s.cleanLine()
	vLine := log.FG_WHITE + log.BG_LBLUE + " " + log.RESET
	fmt.Println("")
	fmt.Printf(vLine + "\tr/c\t- continue viewing statistics\n")
	fmt.Printf(vLine + "\tp\t- pause statistics\n")
	fmt.Printf(vLine + "\tq\t- stop and exit\n")
	fmt.Printf(vLine + "\ta\t- actions\n")
	fmt.Printf(vLine + "\t><\t- view next/preview statistics\n")
	fmt.Printf(vLine + "\tUp/Down\t- sort ascending/descending\n")
	//fmt.Printf("\tl   - limit statistics\n")
	fmt.Printf(vLine + "\tf/F\t- filter statistics (enable/disable)\n")
	fmt.Printf(vLine + "\th\t- help\n")
	fmt.Printf("\n")
	s.cleanLine()
}

func (s *Screen) printActionsMenu() {
	s.cleanLine()
	fmt.Printf("\n\t1. stop firewall\n")
	fmt.Printf("\t2. load firewall\n")
	//fmt.Printf("\t3. change configuration\n")
	//fmt.Printf("\t4. delete rule\n")
	fmt.Printf("\n")
	s.cleanLine()
}

func (s *Screen) printConnectionDetails(con *protocol.Connection) {
	s.cleanLine()
	fmt.Printf("\n  %s - %d:%s -> %s:%d\n", con.Protocol, con.SrcPort, con.SrcIp, con.DstIp, con.DstPort)
	fmt.Printf("\tUser ID:        \t %d\n", con.UserId)
	fmt.Printf("\tInvoked command:\t")
	for _, parg := range con.ProcessArgs {
		fmt.Printf(" %s", parg)
	}
	fmt.Printf("\n")
	fmt.Printf("\tProcess path:    \t%s\n", con.ProcessPath)
	fmt.Printf("\tProcess CWD:     \t%s\n", con.ProcessCwd)
	fmt.Printf("\n")
	s.cleanLine()
}

func (s *Screen) printRule(idx int, rule *storage.Rule) {
	fmt.Printf("  %-4d [%-20s] [%5s] [%s] [%35s]\n", idx, rule.Node, rule.Action, rule.Duration, rule.Name)
}

// print statistics by type (procs, hosts, ports, addrs...)
func (s *Screen) printStats(nodeAddr, what string, hits uint64) {
	lineWidth := int(s.ttyCols) - (8 + 20 + len(what) + 8)
	fmt.Printf("  [%-20s] %-8d - %s%*s\n", nodeAddr, hits, what, lineWidth, " ")
}

// print details of an event
func (s *Screen) printEvent(conn *storage.Connection) {
	eventAction = actionDeny
	eventTime = log.Wrap(fmt.Sprintf("%v", conn.Time), CYAN)
	if conn.RuleAction == api.ActionAllow {
		eventAction = actionAllow
	}

	s.printFormat = s.printFormatPretty
	if s.viewsConf.Style == ViewStylePlain {
		s.printFormat = s.printFormatPlain
		eventAction = conn.RuleAction
		eventTime = conn.Time
	}
	dstHost := conn.DstHost
	if dstHost == "" {
		dstHost = conn.DstIP
	}

	fmt.Printf(s.printFormat,
		eventTime,
		s.viewsConf.Delimiter,
		conn.Node,
		s.viewsConf.Delimiter,
		eventAction,
		s.viewsConf.Delimiter,
		conn.Protocol,
		s.viewsConf.Delimiter,
		conn.UserID,
		s.viewsConf.Delimiter,
		conn.SrcPort,
		s.viewsConf.Delimiter,
		conn.SrcIP,
		s.viewsConf.Delimiter,
		dstHost,
		s.viewsConf.Delimiter,
		conn.DstPort,
		s.viewsConf.Delimiter,
		conn.ProcessPath,
		log.RESET)
}
