package views

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
)

const (
	FG_BLINK  = "\033[5m"
	UNDERLINE = "\033[4;31m"
	CYAN      = "\033[0;49;96m"

	BORDER_TOP_LEFT    = "\u2552"
	BORDER_TOP_RIGHT   = "\u2555"
	BORDER_BOTTOM_LEFT = "\u2558"
	// FIXME: incorrect symbol
	BORDER_BOTTOM_RIGHT = "\u2519"
	BORDER_HORIZONTAL   = "\u2550"
	BORDER_VERTICAL     = "\u2502"
)

// status labels names
const (
	PAUSED = iota
	UPTIME
	RULES
	CONNECTIONS
	DENIES
	EVENTS
	HELP
)

var (
	ttyRows = uint16(80)
	ttyCols = uint16(80)
)

type termSize struct {
	Rows   uint16
	Cols   uint16
	Xpixel uint16
	Ypixel uint16
}

var (
	labels = map[int]string{
		PAUSED:      log.Wrap("[PAUSED] ", log.FG_WHITE+log.BG_YELLOW),
		UPTIME:      log.Wrap("Uptime:", log.FG_WHITE+log.BG_GREEN),
		RULES:       log.Wrap("Rules:", log.FG_WHITE+log.BG_GREEN),
		CONNECTIONS: log.Wrap("Connections:", log.FG_WHITE+log.BG_GREEN),
		DENIES:      log.Wrap("Denies:", log.FG_WHITE+log.BG_GREEN),
		EVENTS:      log.Wrap("Events:", log.FG_WHITE+log.BG_GREEN),
		HELP:        log.Wrap("(h - help, p - pause, q - quit)", log.FG_WHITE+log.BG_LBLUE),
	}

	pendingRules = 0
	labelRules   = labels[RULES]
	labelStatus  = ""
)

func setBlinkingLabel(labelIdx int) {
	switch labelIdx {
	case RULES:
		pendingRules++
		labelRules = FG_BLINK + log.FG_WHITE + log.BG_RED + log.BOLD + fmt.Sprint("!(", pendingRules, ")") + labels[labelIdx]
	}
}

func resetBlinkingLabel(labelIdx int) {
	switch labelIdx {
	case RULES:
		labelRules = labels[labelIdx]
		pendingRules = 0
	}
}

func cleanLine() {
	fmt.Printf("%*s%s\n", ttyCols, " ", log.RESET)
}

func drawBoxTop(width int) {
	cleanLine()
	fmt.Printf(BORDER_TOP_LEFT)
	// unicode chars are 3bytes length
	for w := 0; w < width-6; w++ {
		fmt.Printf(BORDER_HORIZONTAL)
	}
	fmt.Printf(BORDER_TOP_RIGHT + "\n")
}

func drawBoxBottom(width int) {
	fmt.Printf(BORDER_BOTTOM_LEFT)
	for w := 0; w < width-6; w++ {
		fmt.Printf(BORDER_HORIZONTAL)
	}
	fmt.Printf(BORDER_BOTTOM_RIGHT + "\n")
	cleanLine()
}

func questionBox(title, body, buttonBox string) {
	titleLen := utf8.RuneCountInString(title)
	bodyLen := utf8.RuneCountInString(body)
	btnsLen := utf8.RuneCountInString(buttonBox)
	padding := titleLen
	if bodyLen > padding {
		padding = bodyLen
	}
	if btnsLen > padding {
		padding = btnsLen
	}

	padding += 8 // vertical bar * 2 + space

	drawBoxTop(padding)
	// FIXME: vertical lines
	fmt.Printf("%s %s%*s\n", BORDER_VERTICAL, title, padding-titleLen, BORDER_VERTICAL)
	fmt.Printf("%s %s%*s\n", BORDER_VERTICAL, body, padding-bodyLen, BORDER_VERTICAL)
	fmt.Printf("%s %s%*s\n", BORDER_VERTICAL, buttonBox, padding-btnsLen, BORDER_VERTICAL)
	drawBoxBottom(padding)
	cleanLine()
}

func showTopBar(colList []string) {
	topBar := "  "
	for _, col := range colList {
		topBar += fmt.Sprint(log.Wrap(col, log.FG_WHITE+log.BG_GREEN), " ")
	}
	log.Raw(fmt.Sprint(topBar, "\n"))
}

func showStatusBar() {
	stats := config.apiClient.GetLastStats()
	if getPauseStats() {
		labelStatus = labels[PAUSED]
	} else {
		labelStatus = ""
	}
	uptime, _ := time.ParseDuration(fmt.Sprint(stats.Uptime, "s"))
	uptimeLabel := fmt.Sprint(labels[UPTIME], " ", uptime, " ")
	if nodes.Total() > 1 {
		uptimeLabel = ""
	}
	log.Raw(fmt.Sprint(
		labelStatus, // hidden if not paused
		log.Wrap(fmt.Sprint("[", config.View, "]"), log.FG_WHITE+log.BG_LBLUE), " ",
		uptimeLabel,
		labels[CONNECTIONS], " ", nodes.GetStatsSum(0), " ",
		labels[DENIES], " ", nodes.GetStatsSum(1), " ",
		labelRules, " ", nodes.GetStatsSum(2), " ",
		labels[EVENTS], " ", nodes.GetStatsSum(3), " ",
		labels[HELP], " ",
		"\r"))
}

func getTermSize() {
	tSize := &termSize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(tSize)))

	if int(retCode) == -1 {
		return
	}

	ttyRows = tSize.Rows
	ttyCols = tSize.Cols
}

func getTTYSize() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	if termSize, err := cmd.Output(); err == nil {
		fmt.Sscan(string(termSize), &ttyRows, &ttyCols)
	}
}

func resetScreen() {
	getTermSize()
	fmt.Printf("\033[1J\033[f")
}

func RestoreTTY() {
	cmd := exec.Command("stty", "echo")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func nextView() {
	pos, _ := viewList[config.View]
	pos++
	if pos > totalViews {
		pos = 0
	}
	changeView(pos)
}

func prevView() {
	pos, _ := viewList[config.View]
	pos--
	if pos < 0 {
		pos = totalViews
	}
	changeView(pos)
}

func changeView(pos int) {
	config.View = viewNames[pos]
	switch pos {
	case viewList[ViewGeneral]:
		GeneralStats()
	case viewList[ViewHosts],
		viewList[ViewProcs],
		viewList[ViewAddrs],
		viewList[ViewPorts],
		viewList[ViewUsers]:
		StatsByType(viewNames[pos])
	case viewList[ViewRules]:
		RulesList()
	case viewList[ViewNodes]:
		NodesList()
	}
}
