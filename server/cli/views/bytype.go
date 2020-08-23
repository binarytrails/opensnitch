package views

import (
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

// StatsByType shows the latest statistics of the node(s) by type.
func StatsByType(vtype string) {
	waitForStats()
	colWhat := colHost

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
			// TODO sort by hits

			resetScreen()
			showTopBar([]string{colHits, "-", colWhat})
			for what, hits := range vstats {
				if config.Filter != "" {
					if what == config.Filter {
						printStats(what, hits)
					}
				} else {
					printStats(what, hits)
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
