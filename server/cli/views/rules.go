package views

import (
	"fmt"
	"path"
	"strings"
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
	timeoutCanceled := false
	ruleOp := &protocol.Operator{
		Type:    api.RuleSimple,
		Operand: api.FilterByPath,
		Data:    con.ProcessPath,
	}

	procName := path.Base(con.ProcessPath)
	// TODO: uglify rule name
	ruleName := fmt.Sprint(procName, "-", con.Protocol, "-sport", con.SrcPort, "-dport", con.DstPort)
	alertTitle := log.Bold(log.Red(fmt.Sprintf("**** %s is trying to establish a connection ****", procName)))
	alertBody := log.Blue(fmt.Sprint(procName, ": ", con.SrcPort, ":", con.SrcIp, " -> ", con.DstIp, ":", con.DstPort))
	alertButtons := log.Bold(fmt.Sprint(log.Green("✔ Allow (1)"), ", ", log.Red("✘ Deny (2)"), ", ", "Details (3)", ", ", "Options (4)"))

	// TODO: add more options: regexp rule, etc.
	if con.ProcessPath == "" {
		ruleOp.Operand = api.FilterByDstHost
		ruleOp.Data = con.DstHost
		alertTitle = log.Bold(log.Red("  **** New outgoing connection ****  "))
	}

	questionBox(alertTitle, alertBody, alertButtons)

	timeout, _ := time.ParseDuration(defaultRulesTimeout)
	time.AfterFunc(timeout, func() {
		if !timeoutCanceled {
			log.Important("Timeout, default action applied\n")
			setBlinkingLabel(RULES)
			menus.KeyPressedChan <- &menus.KeyEvent{Char: menus.NotAnswered}
		}
	})

	timeoutCanceled, ruleName, ruleOp = askRulesMenu(con, ruleOp)

	config.apiClient.AddNewRule(&protocol.Rule{
		Name:     ruleName,
		Enabled:  true,
		Action:   api.ActionAllow,
		Duration: api.Rule30s,
		Operator: ruleOp,
	})

}

func askRulesMenu(con *protocol.Connection, ruleOp *protocol.Operator) (cancelTimeout bool, ruleName string, newRuleOp *protocol.Operator) {
	newRuleOp = ruleOp
WaitAction:
	printPrompt("")
	switch key := <-menus.KeyPressedChan; key.Char {
	case menus.Allow:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
		resetBlinkingLabel(RULES)

	case menus.Deny:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Red("✘ Connection DENIED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionDeny)
		resetBlinkingLabel(RULES)

	case menus.ShowConnectionDetails:
		printConnectionDetails(con)
		goto WaitAction

	case menus.EditRule:
		newRuleOp = editRule(ruleName, con)
		goto WaitAction

	case menus.NotAnswered:
		// TODO: configure default action
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
	default:
		goto WaitAction
	}
	cancelTimeout = true
	time.Sleep(1 * time.Second)

	pauseStats = false
	return cancelTimeout, ruleName, newRuleOp
}

// TODO: allow to edit more complex rules.
func editRule(ruleName string, con *protocol.Connection) *protocol.Operator {
	filterBy := api.FilterByPath
	filterData := ""
	procArgs := strings.Join(con.ProcessArgs, " ")

WaitAction:
	printConnectionDetails(con)
	fmt.Printf("\n  Filter by:\n")
	fmt.Printf("\t1. process path (%s)\n", con.ProcessPath)
	fmt.Printf("\t2. process command (%s)\n", procArgs)
	fmt.Printf("\t3. destination user id (%d)\n", con.UserId)
	fmt.Printf("\t4. destination port (%d)\n", con.DstPort)
	fmt.Printf("\t5. destination IP (%s)\n", con.DstIp)
	fmt.Printf("\t6. destination host (%s)\n\n", con.DstHost)
	printPrompt("filter by")

	switch key := <-menus.KeyPressedChan; key.Char {
	case menus.FilterByPath:
		filterBy = api.FilterByPath
		filterData = con.ProcessPath
	case menus.FilterByCommand:
		filterBy = api.FilterByCommand
		filterData = procArgs
	case menus.FilterByUserID:
		filterBy = api.FilterByUserID
		filterData = fmt.Sprint(con.UserId)
	case menus.FilterByDstPort:
		filterBy = api.FilterByDstPort
		filterData = fmt.Sprint(con.DstPort)
	case menus.FilterByDstIP:
		filterBy = api.FilterByDstIP
		filterData = con.DstIp
	case menus.FilterByDstHost:
		filterBy = api.FilterByDstHost
		filterData = con.DstHost
	case menus.NotAnswered:
		menus.KeyPressedChan <- &menus.KeyEvent{Char: menus.NotAnswered}
		break
	default:
		goto WaitAction
	}
	return &protocol.Operator{
		Type:    api.RuleSimple,
		Operand: filterBy,
		Data:    filterData,
	}
}
