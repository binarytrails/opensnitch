package views

import (
	"sort"
	"time"

	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
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

type eventHits struct {
	Key   string
	Value uint64
}

func sortEvents(events *map[string]uint64) *[]*eventHits {
	var eves []*eventHits
	for k, v := range *events {
		eves = append(eves, &eventHits{k, v})
	}
	sort.Slice(eves, func(i, j int) bool {
		if sortMode == sortModeAscending {
			return eves[i].Value > eves[j].Value
		}
		return eves[i].Value < eves[j].Value
	})
	return &eves
}

func getTypeStats(viewType string, stype *protocol.Statistics) (vstats *map[string]uint64) {
	vstats = &stype.ByHost
	switch viewType {
	case ViewProcs:
		vstats = &stype.ByExecutable
	case ViewAddrs:
		vstats = &stype.ByAddress
	case ViewPorts:
		vstats = &stype.ByPort
	case ViewUsers:
		vstats = &stype.ByUid
	}

	return vstats
}

func getColumnName(viewType string) (colName string) {
	switch viewType {
	case ViewProcs:
		colName = colProcess
	case ViewAddrs:
		colName = colAddress
	case ViewPorts:
		colName = colPort
	case ViewUsers:
		colName = colUID
	default:
		colName = colHost
	}

	return colName
}

// StatsByType shows the latest statistics of the node(s) by type.
func StatsByType(viewType string) {
	waitForStats()
	colWhat := getColumnName(viewType)
	topCols := []string{"Node                  ", colHits, "-", colWhat}
	totalStats := 0

	for {
		if !getPauseStats() {
			resetScreen()
			topCols[3] = colWhat
			showTopBar(topCols)

			totalStats = 0
			for addr, node := range *nodes.GetAll() {
				if node.GetStats() == nil {
					break
				}

				vstats := getTypeStats(viewType, node.GetStats())
				for _, e := range *sortEvents(vstats) {
					// XXX: actions should be modular
					if config.Filter != "" {
						if e.Key == config.Filter {
							printStats(addr, e.Key, e.Value)
						}
					} else {
						printStats(addr, e.Key, e.Value)
					}
				}

				totalStats += len(*vstats)
			}
			printVerticalPadding(totalStats)
		}
		showStatusBar()
		readLiveMenu()
		if getStopStats() {
			return
		}
		time.Sleep(600 * time.Millisecond)
	}
}
