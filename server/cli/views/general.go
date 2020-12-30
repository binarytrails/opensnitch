package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/gustavo-iniguez-goya/opensnitch/server/api/storage"
)

// ViewEvents holds the functionality to display connections events
type ViewEvents struct {
	*Screen
	*BaseView
}

// NewViewEvents returns a ViewEvents struct and initializes the parent structs.
func NewViewEvents(scr *Screen, baseView *BaseView) *ViewEvents {
	return &ViewEvents{scr, baseView}
}

// Print displays the latest statistics of the node(s).
func (v *ViewEvents) Print() {
	v.waitForStats()

	totalEvents := 0
	for {
		if !v.getPauseStats() {
			v.resetScreen()

			events := v.aClient.GetEvents(v.sortMode, v.viewsConf.Limit)
			totalEvents = len(*events)
			for _, event := range *events {
				if v.viewsConf.Filter != "" {
					v.filterEvent(&event)
				} else {
					v.printEvent(&event)
				}
			}
			v.printVerticalPadding(totalEvents)
		}
		v.showStatusBar()
		readLiveMenu()
		if ui.getStopStats() {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func (v *ViewEvents) filterEvent(conn *storage.Connection) {
	switch v.viewsConf.Filter {
	case conn.Protocol:
	case fmt.Sprint(conn.UserID):
	case fmt.Sprint(conn.SrcPort):
	case conn.SrcIP:
	case fmt.Sprint(conn.DstPort):
	case conn.DstIP:
	case conn.DstHost:
	case conn.RuleAction:
		screen.printEvent(conn)
	}
	if strings.Contains(conn.ProcessPath, v.viewsConf.Filter) ||
		strings.Contains(conn.DstHost, v.viewsConf.Filter) {
		screen.printEvent(conn)
	}
}
