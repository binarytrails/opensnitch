package views

import (
	"fmt"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"time"
)

// ViewNodes holds the functionality to list nodes.
type ViewNodes struct {
	*Screen
	*BaseView
}

// NewViewNodes returns a new ViewNodes struct and initializes the parent structs.
func NewViewNodes(scr *Screen, baseView *BaseView) *ViewNodes {
	return &ViewNodes{scr, baseView}
}

// Print displays the list of connected nodes.
func (v *ViewNodes) Print() {
	v.waitForStats()
	topCols := []string{"Last seen        ", " - ", "Node                  ", " - ", "Status   ", " - ", "Version                              ", " - ", "Name     "}
	for {
		if !v.getPauseStats() {
			v.resetScreen()
			v.showTopBar(topCols)
			for _, node := range *nodes.GetAll() {
				fmt.Printf("  %v\n", node)
			}
			v.printVerticalPadding(nodes.Total())
		}
		if v.getStopStats() {
			return
		}
		v.showStatusBar()
		readLiveMenu()
		time.Sleep(300 * time.Millisecond)
	}
}
