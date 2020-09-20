package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
)

// GeneralStats displays the latest statistics of the node(s).
func GeneralStats() {
	waitForStats()

	totalEvents := 0
	for {
		if !getPauseStats() {
			resetScreen()

			allEvents := collectEvents()
			totalEvents = len(allEvents)
			for idx, event := range allEvents {
				if idx == config.Limit {
					break
				}
				if config.Filter != "" {
					filterEvent(event)
				} else {
					printEvent(event)
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

// collect events from all the connected nodes.
func collectEvents() (events []*protocol.Event) {
	for _, node := range *nodes.GetAll() {
		if node.GetStats() == nil {
			continue
		}
		events = append(events, node.GetStats().Events...)
	}
	sort.Slice(events, func(i, j int) bool {
		if sortMode == sortModeAscending {
			return events[i].Time < events[j].Time
		}
		return events[i].Time > events[j].Time
	})

	return events
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
