package views

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/storage"
	"github.com/gustavo-iniguez-goya/opensnitch/server/cli/menus"
)

var missedEvents []*storage.Connection

// ViewRules holds the information and functionality to work with rules
type ViewRules struct {
	*Screen
	*BaseView
}

// NewViewRules returns a new ViewRules struct and initializes the parent structs.
func NewViewRules(scr *Screen, baseView *BaseView) *ViewRules {
	return &ViewRules{scr, baseView}
}

// Print lists all the rules the nodes have.
func (v *ViewRules) Print() {
	v.waitForStats()
	v.resetBlinkingLabel(RULES)
	topCols := []string{"Num ", "Node                  ", "Action ", "Duration", "Name                                 "}
	rulNums := 0
	missedLabel := log.Red("  MISSED EVENTS")

	for {
		if !v.getPauseStats() {
			v.resetScreen()
			v.showTopBar(topCols)
			rules := v.aClient.GetRules(v.sortMode, v.viewsConf.Limit)
			rulNums = len(*rules)
			for idx, rule := range *rules {
				v.printRule(idx, &rule)
			}
			if len(missedEvents) > 0 {
				fmt.Printf("  %s\n", missedLabel)
				for _, conn := range missedEvents {
					rulNums++
					v.printEvent(conn)
				}
			}
			v.printVerticalPadding(rulNums)
		}
		if v.getStopStats() {
			return
		}
		v.showStatusBar()
		readLiveMenu()
		time.Sleep(300 * time.Millisecond)
	}
}

func (v *ViewRules) askRule(con *protocol.Connection) {
	lock.Lock()
	defer lock.Unlock()
	v.pauseStats = true
	timeoutCanceled := false
	procName := path.Base(con.ProcessPath)
	ruleName := fmt.Sprint(procName, "-", con.Protocol, "-sport", con.SrcPort, "-dport", con.DstPort)

	ruleOp := &protocol.Operator{
		Type:    api.RuleSimple,
		Operand: api.FilterByPath,
		Data:    con.ProcessPath,
	}
	rule := &protocol.Rule{
		Name:     ruleName,
		Enabled:  true,
		Action:   api.ActionAllow,
		Duration: api.Rule30s,
		Operator: ruleOp,
	}

	// TODO: uglify rule name
	alertTitle := log.Bold(log.Red(fmt.Sprintf("**** %s is trying to establish a connection ****", procName)))
	alertBody := log.Blue(fmt.Sprint(procName, ": ", con.SrcPort, ":", con.SrcIp, " -> ", con.DstIp, ":", con.DstPort))
	alertButtons := log.Bold(fmt.Sprint(log.Green("✔ Allow (1)"), ", ", log.Red("✘ Deny (2)"), ", ", "Details (3)", ", ", "Options (4)"))

	// TODO: add more options: regexp rule, etc.
	if con.ProcessPath == "" {
		ruleOp.Operand = api.FilterByDstHost
		ruleOp.Data = con.DstHost
		alertTitle = log.Bold(log.Red("  **** New outgoing connection ****  "))
	}

	v.questionBox(alertTitle, alertBody, alertButtons)

	timeout, _ := time.ParseDuration(defaultRulesTimeout)
	time.AfterFunc(timeout, func() {
		if !timeoutCanceled {
			log.Important("Timeout, default action applied\n")
			v.setBlinkingLabel(RULES)
			menus.KeyPressedChan <- &menus.KeyEvent{Char: menus.NotAnswered}
		}
	})

	timeoutCanceled, ruleName = v.askRulesMenu(con, rule)
	v.aClient.AddNewRule(rule)
}

func (v *ViewRules) askRulesMenu(con *protocol.Connection, rule *protocol.Rule) (cancelTimeout bool, ruleName string) {
WaitAction:
	v.printPrompt("")
	switch key := <-menus.KeyPressedChan; key.Char {
	case menus.Allow:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), v.ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
		v.resetBlinkingLabel(RULES)

	case menus.Deny:
		fmt.Printf("%s%*s\n\n", log.Bold(log.Red("✘ Connection DENIED")), v.ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionDeny)
		v.resetBlinkingLabel(RULES)

	case menus.ShowConnectionDetails:
		v.printConnectionDetails(con)
		goto WaitAction

	case menus.EditRule:
		v.editRule(ruleName, con, rule)
		fmt.Printf("  %s - %s\n", rule.Operator.Type, rule.Operator.Data)
		goto WaitAction

	case menus.NotAnswered:
		// TODO: configure default action
		fmt.Printf("%s%*s\n\n", log.Bold(log.Green("✔ Connection ALLOWED")), v.ttyCols, " ")
		ruleName = fmt.Sprint(ruleName, api.ActionAllow)
		v.addMissedEvent(con, rule)
	default:
		goto WaitAction
	}
	cancelTimeout = true
	time.Sleep(1 * time.Second)

	v.pauseStats = false
	return cancelTimeout, ruleName
}

// TODO: allow to edit more complex rules.
func (v *ViewRules) editRule(ruleName string, con *protocol.Connection, rule *protocol.Rule) {
	filterBy := api.FilterByPath
	filterData := ""
	procArgs := strings.Join(con.ProcessArgs, " ")

WaitAction:
	v.printConnectionDetails(con)
	fmt.Printf("\n  Filter by:\n")
	fmt.Printf("\t1. process path (%s)\n", con.ProcessPath)
	fmt.Printf("\t2. process command (%s)\n", procArgs)
	fmt.Printf("\t3. destination user id (%d)\n", con.UserId)
	fmt.Printf("\t4. destination port (%d)\n", con.DstPort)
	fmt.Printf("\t5. destination IP (%s)\n", con.DstIp)
	fmt.Printf("\t6. destination host (%s)\n\n", con.DstHost)
	v.printPrompt("filter by")

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
	rule.Operator.Type = api.RuleSimple
	rule.Operator.Operand = filterBy
	rule.Operator.Data = filterData

}

func (v *ViewRules) addMissedEvent(con *protocol.Connection, rule *protocol.Rule) {
	missedEvents = append(missedEvents,
		&storage.Connection{
			Node:        "<xxx>",
			RuleName:    rule.Name,
			Time:        time.Now().Format("2006/01/02 00:01:02"),
			Protocol:    con.Protocol,
			SrcIP:       con.SrcIp,
			SrcPort:     con.SrcPort,
			DstIP:       con.DstIp,
			DstHost:     con.DstHost,
			DstPort:     con.DstPort,
			UserID:      con.UserId,
			PID:         con.ProcessId,
			ProcessPath: con.ProcessPath,
			ProcessCwd:  con.ProcessCwd,
			ProcessArgs: strings.Join(con.ProcessArgs, " "),
		})
}
