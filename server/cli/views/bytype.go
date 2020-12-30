package views

import (
	"sort"
	"time"

	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
)

const (
	// hits column is 10chars wide
	colHits    = "Hits    "
	colHost    = "Host"
	colProcess = "Process"
	colAddress = "Address"
	colPort    = "Port"
	colUID     = "UID"
)

// ViewEventsByType holds the functionality to list events by type.
type ViewEventsByType struct {
	*Screen
	*BaseView
	eventType map[int8]string
}

// NewViewByType returns a new ViewByType struct and initializes the parent structs.
func NewViewByType(scr *Screen, baseView *BaseView) *ViewEventsByType {
	return &ViewEventsByType{
		scr,
		baseView,
		map[int8]string{
			0: "ByProto",
			1: "ByHost",
			2: "ByExecutable",
			3: "ByAddress",
			4: "ByPort",
			5: "ByUid",
		},
	}
}

type eventHits struct {
	Key   string
	Value uint64
}

func (v *ViewEventsByType) sortEvents(events *map[string]uint64) *[]*eventHits {
	var eves []*eventHits
	for k, v := range *events {
		eves = append(eves, &eventHits{k, v})
	}
	sort.Slice(eves, func(i, j int) bool {
		if v.sortMode == v.sortModeAscending {
			return eves[i].Value > eves[j].Value
		}
		return eves[i].Value < eves[j].Value
	})
	return &eves
}

func (v *ViewEventsByType) getTypeStats(viewName string, stype *protocol.Statistics) (vstats *map[string]uint64) {
	vstats = &stype.ByHost
	switch viewName {
	case ViewNameProcs:
		vstats = &stype.ByExecutable
	case ViewNameAddrs:
		vstats = &stype.ByAddress
	case ViewNamePorts:
		vstats = &stype.ByPort
	case ViewNameUsers:
		vstats = &stype.ByUid
	}

	return vstats
}

func (v *ViewEventsByType) getColumnName(viewName string) (colName string) {
	switch viewName {
	case ViewNameProcs:
		colName = colProcess
	case ViewNameAddrs:
		colName = colAddress
	case ViewNamePorts:
		colName = colPort
	case ViewNameUsers:
		colName = colUID
	default:
		colName = colHost
	}

	return colName
}

// Print shows the latest statistics of the node(s) by type.
func (v *ViewEventsByType) Print(viewPos int8, viewName string) {
	v.waitForStats()
	colWhat := v.getColumnName(viewName)
	topCols := []string{"Node                  ", colHits, "-", colWhat}
	totalStats := 0

	for {
		if !v.getPauseStats() {
			v.resetScreen()
			topCols[3] = colWhat
			v.showTopBar(topCols)

			totalStats = 0
			vstats := v.aClient.GetEventsByType(v.eventType[viewPos], v.sortMode, v.viewsConf.Limit)
			for _, ev := range *vstats {
				if v.viewsConf.Filter != "" {
					v.printStats(ev.Node, ev.What, ev.Hits)
				} else {
					v.printStats(ev.Node, ev.What, ev.Hits)
				}
			}

			totalStats += len(*vstats)
			v.printVerticalPadding(totalStats)
		}
		v.showStatusBar()
		readLiveMenu()
		if v.getStopStats() {
			return
		}
		time.Sleep(600 * time.Millisecond)
	}
}
