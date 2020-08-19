package menus

import (
	"bufio"
	"os"
	"sync"

	"github.com/eiannone/keyboard"
)

var (
	lock            sync.RWMutex
	stopInteractive = false
	scanner         = bufio.NewScanner(os.Stdin)
	KeyPressedChan  = make(chan string, 1)
)

// menus related constants
const (
	NotAnswered = "0"
	Allow       = "1"
	Deny        = "2"

	PAUSE    = "p"
	RUN      = "r"
	SAVE     = "s"
	CONTINUE = "c"
	HELP     = "h"
	QUIT     = "q"
	NEXTVIEW = ">"
	PREVVIEW = "<"
	LIMIT    = "l"
	SORT     = "o"
	FILTER   = "f"
)

// ReadKey reads a key from the keyboard
func ReadKey() string {
	c, _, err := keyboard.GetSingleKey()
	if err != nil {
		return ""
	}
	return string(c)
}

// WaitForKey pauses execution until a key is pressed
func WaitForKey() string {
	scanner.Scan()
	return scanner.Text()
}

// StopInteractive stops reading and handling keystrokes
func StopInteractive() {
	stopInteractive = true
}

func shouldStop() bool {
	lock.RLock()
	defer lock.RUnlock()

	return stopInteractive
}

// Interactive listens in background for keystrokes.
func Interactive() chan string {
	go func() {
		for {
			if shouldStop() {
				return
			}
			KeyPressedChan <- ReadKey()
		}
	}()
	return KeyPressedChan
}

// Exit stop reading keystrokes.
func Exit() {
	lock.Lock()
	defer lock.Unlock()

	StopInteractive()
	keyboard.Close()
}
