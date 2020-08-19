package views

import (
	"time"
)

// GeneralStats displays the latest statistics of the node(s).
func GeneralStats() {
	waitForStats()

	for {
		if !getPauseStats() {
			resetScreen()
			totalEvents := len(config.apiClient.GetLastStats().Events)
			limit := 0
			if totalEvents >= config.Limit {
				limit = totalEvents - config.Limit
			}
			// TODO: sortEvents()
			// TODO: filter by field(s), limit
			for idx := limit; idx < totalEvents; idx++ {
				printEvent(config.apiClient.GetLastStats().Events[idx])
			}
		}
		showStatusBar()
		readLiveMenu()
		if getStopStats() || !config.Loop {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}
