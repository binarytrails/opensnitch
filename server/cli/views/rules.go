package views

import (
	"fmt"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"time"
)

// RulesList lists all the rules the nodes have.
func RulesList() {
	waitForStats()
	rules := make(map[string]*protocol.Rule)
	for {
		if !getPauseStats() {
			resetScreen()
			showTopBar([]string{"Num ", " - ", "Node                  ", "Name                                 ", "Action ", "Duration"})
			for addr, node := range *nodes.GetAll() {
				for idx, rule := range node.GetConfig().Rules {
					if _, found := rules[rule.Name]; !found {
						fmt.Printf("  %-5d -  [%-20s] [%35s] [%5s] [%s]\n", idx, addr, rule.Name, rule.Action, rule.Duration)
					}
				}
			}
		}
		if getStopStats() || !config.Loop {
			return
		}
		showStatusBar()
		readLiveMenu()
		time.Sleep(300 * time.Millisecond)
	}
}
