package views

import (
	"fmt"
	"path"
	"time"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"github.com/gustavo-iniguez-goya/opensnitch/server/cli/menus"
)

// RulesList lists all the rules the nodes have.
func RulesList() {
	waitForStats()
	rules := make(map[string]*protocol.Rule)
	topCols := []string{"Num ", "Node                  ", "Name                                 ", "Action ", "Duration"}
	rulNums := 0
	for {
		if !getPauseStats() {
			resetScreen()
			showTopBar(topCols)
			rulNums = 0
			for addr, node := range *nodes.GetAll() {
				for idx, rule := range node.GetConfig().Rules {
					rulNums++
					if _, found := rules[rule.Name]; !found {
						fmt.Printf("  %-4d [%-20s] [%35s] [%5s] [%s]\n", idx, addr, rule.Name, rule.Action, rule.Duration)
					}
				}
			}
			printVerticalPadding(rulNums)
		}
		if getStopStats() {
			return
		}
		showStatusBar()
		readLiveMenu()
		time.Sleep(300 * time.Millisecond)
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
		menus.KeyPressedChan <- &menus.KeyEvent{Char: menus.NotAnswered}
	})
	printConnectionDetails(con)

	cleanLine()
WaitAction:
	switch key := <-menus.KeyPressedChan; key.Char {
	case menus.Allow:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
		resetBlinkingLabel(RULES)
	case menus.Deny:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Red("✘ Connection DENIED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionDeny)
		resetBlinkingLabel(RULES)
	// TODO
	case menus.ShowConnectionDetails:
		pauseStats = true
		printConnectionDetails(con)
		pauseStats = false
	case menus.NotAnswered:
		// TODO: configure default action
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
	default:
		goto WaitAction
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
