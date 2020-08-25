package menus

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/eiannone/keyboard"
)

// KeyEvent is the keystroke pressed by the user
type KeyEvent struct {
	Key  keyboard.Key
	Char string
}

var (
	lock            sync.RWMutex
	stopInteractive = false
	scanner         = bufio.NewScanner(os.Stdin)
	// KeyPressedChan is the channel where the keystrokes are sent/received.
	KeyPressedChan = make(chan *KeyEvent, 1)
)

// menus related constants
const (
	// rules
	NotAnswered           = "0"
	Allow                 = "1"
	Deny                  = "2"
	ShowConnectionDetails = "3"
	EditRule              = "4"
	FilterByPath          = "1"
	FilterByCommand       = "2"
	FilterByUserID        = "3"
	FilterByDstPort       = "4"
	FilterByDstIP         = "5"
	FilterByDstHost       = "6"

	PAUSE          = "p"
	RUN            = "r"
	SAVE           = "s"
	CONTINUE       = "c"
	HELP           = "h"
	QUIT           = "q"
	NEXTVIEW       = ">"
	NEXTVIEWARROW  = keyboard.KeyArrowRight
	PREVVIEW       = "<"
	PREVVIEWARROW  = keyboard.KeyArrowLeft
	LIMIT          = "l"
	SORT           = "o"
	SORTASCENDING  = keyboard.KeyArrowUp
	SORTDESCENDING = keyboard.KeyArrowDown
	FILTER         = "f"
	DISABLEFILTER  = "F"
	NOTIFICATIONS  = "n"
	ACTIONS        = "a"

	STOPFIREWALL  = "1"
	STARTFIREWALL = "2"
	CHANGECONFIG  = "3"
	DELETERULE    = "4"
)

// ReadKey reads a key from the keyboard
func ReadKey() *KeyEvent {
	c, key, err := keyboard.GetSingleKey()
	if err != nil {
		return nil
	}
	return &KeyEvent{
		Key:  key,
		Char: string(c),
	}
}

// ReadLine reads a sequence of keystrokes until Enter is pressed.
func ReadLine() (str string) {
	for {
		key := <-KeyPressedChan
		switch {
		case key.Key == keyboard.KeyEnter, key.Key == keyboard.KeyEsc:
			goto Exit
		default:
			fmt.Printf("%s", key.Char)
			str += key.Char
		}
	}
Exit:

	return str
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
func Interactive() chan *KeyEvent {
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
