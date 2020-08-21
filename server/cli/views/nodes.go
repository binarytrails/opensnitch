package views

import (
	"fmt"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"time"
)

// NodesList displays the list of connected nodes.
func NodesList() {
	waitForStats()
	for {
		if !getPauseStats() {
			resetScreen()
			showTopBar([]string{"Total", " - ", "Node    ", "Name     ", "Version"})
			for addr, node := range *nodes.GetAll() {
				fmt.Printf("%-4d - [%s] [%s] [%s]\n", nodes.Total(), addr, node.GetConfig().Name, node.GetConfig().Version)
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
