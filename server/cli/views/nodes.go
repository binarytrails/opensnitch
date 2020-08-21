package views

import (
	"fmt"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"time"
)

// NodesList displays the list of connected nodes.
func NodesList() {
	waitForStats()
	topCols := []string{"Last seen        ", " - ", "Node                 ", " - ", "Status   ", " - ", "Version                             ", " - ", "Name     "}
	for {
		if !getPauseStats() {
			resetScreen()
			showTopBar(topCols)
			for addr, node := range *nodes.GetAll() {
				fmt.Printf("  [%v]  - [%-20s]  -  [%-25s] - [%s]  -  [%s]\n", node.LastSeen().Format(time.Stamp), addr, node.Status(), node.GetConfig().Version, node.GetConfig().Name)
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
