package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/ui/protocol"
)

// GeneralStats displays the latest statistics of the node(s).
func GeneralStats() {
	waitForStats()

	limit := 0
	totalEvents := 0
	for {
		if !getPauseStats() {
			resetScreen()
			totalEvents = len(config.apiClient.GetLastStats().Events)
			limit = 0
			if totalEvents >= config.Limit {
				limit = totalEvents - config.Limit
			}
			// TODO: sortEvents()
			for idx := limit; idx < totalEvents; idx++ {
				if config.Filter != "" {
					filterEvent(config.apiClient.GetLastStats().Events[idx])
				} else {
					printEvent(config.apiClient.GetLastStats().Events[idx])
				}
			}
			printVerticalPadding(totalEvents)
		}
		showStatusBar()
		readLiveMenu()
		if getStopStats() || !config.Loop {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func filterEvent(ev *protocol.Event) {
	switch config.Filter {
	case ev.Connection.Protocol:
	case fmt.Sprint(ev.Connection.UserId):
	case fmt.Sprint(ev.Connection.SrcPort):
	case ev.Connection.SrcIp:
	case fmt.Sprint(ev.Connection.DstPort):
	case ev.Connection.DstIp:
	case ev.Connection.DstHost:
	case ev.Rule.Action:
		printEvent(ev)
	}
	if strings.Contains(ev.Connection.ProcessPath, config.Filter) {
		printEvent(ev)
	}
}
