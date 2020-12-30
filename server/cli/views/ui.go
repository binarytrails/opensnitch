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
	"github.com/gustavo-iniguez-goya/opensnitch/server/api"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/storage"
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

type termSize struct {
	Rows   uint16
	Cols   uint16
	Xpixel uint16
	Ypixel uint16
}

// UI holds the functionality about the UI, and how
// the data is presented to the user.
type UI struct {
	aClient   *api.Client
	viewsConf *Config

	labels       map[int]string
	pendingRules int
	labelRules   string
	labelStatus  string

	stopStats  bool
	pauseStats bool

	ttyRows uint16
	ttyCols uint16
}

// NewUI returns a new UI struct and initializes default values.
func NewUI(aClient *api.Client, conf *Config) *UI {
	ui := &UI{
		viewsConf: conf,
		aClient:   aClient,
		labels: map[int]string{
			PAUSED:      log.Wrap("[PAUSED] ", log.FG_WHITE+log.BG_YELLOW),
			UPTIME:      log.Wrap("Uptime:", log.FG_WHITE+log.BG_GREEN),
			RULES:       log.Wrap("Rules:", log.FG_WHITE+log.BG_GREEN),
			CONNECTIONS: log.Wrap("Connections:", log.FG_WHITE+log.BG_GREEN),
			DENIES:      log.Wrap("Denies:", log.FG_WHITE+log.BG_GREEN),
			EVENTS:      log.Wrap("Events:", log.FG_WHITE+log.BG_GREEN),
			HELP:        log.Wrap("(h - help, p - pause, q - quit)", log.FG_WHITE+log.BG_LBLUE),
		},
	}

	ui.pendingRules = 0
	ui.labelRules = ui.labels[RULES]
	ui.labelStatus = ""

	ui.getTermSize()

	return ui
}

func (u *UI) waitForStats() {
	tries := 30
	for u.aClient.GetLastStats() == nil && tries > 0 {
		time.Sleep(1 * time.Second)
		tries--
		log.Raw(log.Wrap("No stats yet, waiting ", log.GREEN)+" %d\r", tries)
	}
	u.cleanLine()
}

func (u *UI) getGlobalStats() *storage.Statistics {
	stats := &storage.Statistics{}
	dbStats := u.aClient.GetStats()

	if len(*dbStats) == 1 {
		stats.Uptime = (*dbStats)[0].Uptime
	}

	for _, s := range *dbStats {
		stats.Rules += s.Rules
		stats.DNSResponses += s.DNSResponses
		stats.Connections += s.Connections
		stats.Dropped += s.Dropped
		stats.RuleHits += s.RuleHits
		stats.RuleMisses += s.RuleMisses
	}

	return stats
}

func (u *UI) getPauseStats() bool {
	return u.pauseStats
}

func (u *UI) getStopStats() bool {
	return u.stopStats || !u.viewsConf.Loop
}

func (u *UI) setBlinkingLabel(labelIdx int) {
	switch labelIdx {
	case RULES:
		ui.pendingRules++
		ui.labelRules = FG_BLINK + log.FG_WHITE + log.BG_RED + log.BOLD + fmt.Sprint("!(", ui.pendingRules, ")") + ui.labels[labelIdx]
	}
}

func (u *UI) resetBlinkingLabel(labelIdx int) {
	switch labelIdx {
	case RULES:
		ui.labelRules = ui.labels[labelIdx]
		ui.pendingRules = 0
	}
}

func (u *UI) cleanLine() {
	fmt.Printf("%*s%s\n", u.ttyCols, " ", log.RESET)
}

func (u *UI) drawBoxTop(width int) {
	u.cleanLine()
	fmt.Printf(BORDER_TOP_LEFT)
	// unicode chars are 3bytes length
	for w := 0; w < width-6; w++ {
		fmt.Printf(BORDER_HORIZONTAL)
	}
	fmt.Printf(BORDER_TOP_RIGHT + "\n")
}

func (u *UI) drawBoxBottom(width int) {
	fmt.Printf(BORDER_BOTTOM_LEFT)
	for w := 0; w < width-6; w++ {
		fmt.Printf(BORDER_HORIZONTAL)
	}
	fmt.Printf(BORDER_BOTTOM_RIGHT + "\n")
	u.cleanLine()
}

func (u *UI) questionBox(title, body, buttonBox string) {
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

	u.drawBoxTop(padding)
	// FIXME: vertical lines
	fmt.Printf("%s %s%*s\n", BORDER_VERTICAL, title, padding-titleLen, BORDER_VERTICAL)
	fmt.Printf("%s %s%*s\n", BORDER_VERTICAL, body, padding-bodyLen, BORDER_VERTICAL)
	fmt.Printf("%s %s%*s\n", BORDER_VERTICAL, buttonBox, padding-btnsLen, BORDER_VERTICAL)
	u.drawBoxBottom(padding)
	u.cleanLine()
}

func (u *UI) showTopBar(colList []string) {
	topBar := "  "
	for _, col := range colList {
		topBar += fmt.Sprint(log.Wrap(col, log.FG_WHITE+log.BG_GREEN), " ")
	}
	log.Raw(fmt.Sprint(topBar, "\n"))
}

func (u *UI) showStatusBar() {
	//stats := u.aClient.GetLastStats()
	stats := u.getGlobalStats()
	if u.getPauseStats() {
		ui.labelStatus = ui.labels[PAUSED]
	} else {
		ui.labelStatus = ""
	}
	uptime, _ := time.ParseDuration(fmt.Sprint(stats.Uptime, "s"))
	uptimeLabel := fmt.Sprint(ui.labels[UPTIME], " ", uptime, " ")
	if nodes.Total() > 1 {
		uptimeLabel = ""
	}
	log.Raw(fmt.Sprint(
		ui.labelStatus, // hidden if not paused
		log.Wrap(fmt.Sprint("[", u.viewsConf.View, "]"), log.FG_WHITE+log.BG_LBLUE), " ",
		uptimeLabel,
		ui.labels[CONNECTIONS], " ", nodes.GetStatsSum(0), " ",
		ui.labels[DENIES], " ", nodes.GetStatsSum(1), " ",
		ui.labelRules, " ", nodes.GetStatsSum(2), " ",
		ui.labels[EVENTS], " ", nodes.GetStatsSum(3), " ",
		ui.labels[HELP], " ",
		"\r"))
}

func (u *UI) getTermSize() {
	tSize := &termSize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(tSize)))

	if int(retCode) == -1 {
		return
	}

	u.ttyRows = tSize.Rows
	u.ttyCols = tSize.Cols
}

func getTTYSize() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	if termSize, err := cmd.Output(); err == nil {
		fmt.Sscan(string(termSize), &ui.ttyRows, &ui.ttyCols)
	}
}

func (u *UI) resetScreen() {
	u.getTermSize()
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
	if pos >= totalViews {
		pos = 0
	}
	changeView(pos)
}

func prevView() {
	pos, _ := viewList[config.View]
	pos--
	if pos < 0 {
		pos = totalViews - 1
	}
	changeView(pos)
}

func changeView(pos int8) {
	println("POS:", pos)
	config.View = viewNames[pos]
	switch pos {
	case viewList[ViewNameGeneral]:
		vEvents.Print()
	case viewList[ViewNameHosts],
		viewList[ViewNameProcs],
		viewList[ViewNameAddrs],
		viewList[ViewNamePorts],
		viewList[ViewNameUsers]:
		vEventsByType.Print(pos, viewNames[pos])
	case viewList[ViewNameRules]:
		vRules.Print()
	case viewList[ViewNameNodes]:
		vNodes.Print()
	}
}
