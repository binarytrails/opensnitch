package views

import (
	"fmt"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"time"
)

// NodesList displays the list of connected nodes.
func NodesList() {
	waitForStats()
	topCols := []string{"Last seen        ", " - ", "Node                  ", " - ", "Status   ", " - ", "Version                              ", " - ", "Name     "}
	for {
		if !getPauseStats() {
			resetScreen()
			showTopBar(topCols)
			for _, node := range *nodes.GetAll() {
				fmt.Printf("  %v\n", node)
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
