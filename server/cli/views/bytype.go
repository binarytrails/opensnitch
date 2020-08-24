package views

import (
	"sort"
	"time"
)

const (
	// hits column is 5chars width
	colHits    = "Hits "
	colHost    = "Host"
	colProcess = "Process"
	colAddress = "Address"
	colPort    = "Port"
	colUID     = "UID"
)

type event struct {
	Key   string
	Value uint64
}

func sortEvents(events map[string]uint64) *[]*event {
	var eves []*event
	for k, v := range events {
		eves = append(eves, &event{k, v})
	}
	sort.Slice(eves, func(i, j int) bool {
		if sortMode == sortModeAscending {
			return eves[i].Value > eves[j].Value
		}
		return eves[i].Value < eves[j].Value
	})
	return &eves
}

// StatsByType shows the latest statistics of the node(s) by type.
func StatsByType(vtype string) {
	waitForStats()
	colWhat := colHost
	topCols := []string{colHits, "-", colWhat}

	for {
		if !getPauseStats() {
			vstats := config.apiClient.GetLastStats().ByHost
			switch vtype {
			case ViewProcs:
				vstats = config.apiClient.GetLastStats().ByExecutable
				colWhat = colProcess
			case ViewAddrs:
				vstats = config.apiClient.GetLastStats().ByAddress
				colWhat = colAddress
			case ViewPorts:
				vstats = config.apiClient.GetLastStats().ByPort
				colWhat = colPort
			case ViewUsers:
				vstats = config.apiClient.GetLastStats().ByUid
				colWhat = colUID
			}
			resetScreen()
			topCols[2] = colWhat
			showTopBar(topCols)

			for _, e := range *sortEvents(vstats) {
				if config.Filter != "" {
					if e.Key == config.Filter {
						printStats(e.Key, e.Value)
					}
				} else {
					printStats(e.Key, e.Value)
				}
			}
			printVerticalPadding(len(vstats))
		}
		showStatusBar()
		readLiveMenu()
		if getStopStats() {
			return
		}
		time.Sleep(600 * time.Millisecond)
	}
}
