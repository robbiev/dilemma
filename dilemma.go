package dilemma

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	// Empty means key code information is not applicable
	Empty Key = iota
	up
	down
	enter
	// CtrlC means CTRL-C was pressed.
	// Usually this means the user wants to send SIGINT.
	CtrlC
)

const (
	exitNo exitStatus = iota
	exitYes
)

const (
	helpNo helpStatus = iota
	helpYes
)

// Key represents keys pressed by the user.
type Key int

type exitStatus int

type helpStatus int

// Config holds the configuration to display a list of options
// for a user to select.
type Config struct {
	Title   string
	Options []string
	Help    string
}

func invertColours() {
	fmt.Print("\033[7m")
}

func resetStyle() {
	fmt.Print("\033[0m")
}

func moveUp() {
	fmt.Print("\033[1A")
}

func clearLine() {
	fmt.Print("\033[2K\r")
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func lineCount(s string) int {
	var count int
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			count++
		}
	}
	return count + 1 // also count the first line
}

func inputLoop(keyPresses chan<- Key, exitAck chan exitStatus) {
	buf := make([]byte, 128)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			panic(err)
		}
		input := string(buf[:n])
		switch {
		case input == "\033[A":
			keyPresses <- up
		case input == "\033[B":
			keyPresses <- down
		case input == "\x0D":
			keyPresses <- enter
		case input == "\x03":
			keyPresses <- CtrlC
		default:
			keyPresses <- Empty
		}
		if exitYes == <-exitAck {
			return
		}
	}
}

// Prompt asks the user to select an option from the list. The selected option
// is returned in the first return value. The second return value is set to
// Empty unless the user presses CTRL-C (indicating she wants to signal SIGINT)
// in which case the value will be CtrlC.
func Prompt(config Config) (selected string, exitKey Key) {
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(0, oldState)

	hideCursor()
	defer showCursor()

	// ensure we always exit with the cursor at the beginning of the line so the
	// terminal prompt prints in the expected place
	defer func() {
		fmt.Print("\r")
	}()

	keyPresses := make(chan Key)
	exitAck := make(chan exitStatus)
	go inputLoop(keyPresses, exitAck)

	var selectionIndex int

	draw := func(help helpStatus) {
		fmt.Println(config.Title)
		fmt.Print("\r")
		for i, v := range config.Options {
			fmt.Print("  ")
			if i == selectionIndex {
				invertColours()
			}
			fmt.Printf("%s\n", v)
			if i == selectionIndex {
				resetStyle()
			}
			fmt.Print("\r")
		}
		if help == helpYes {
			fmt.Print(config.Help)
		}
	}

	clear := func(help helpStatus) {
		lines := lineCount(config.Title) + len(config.Options)

		if help == helpYes {
			lines = lines + lineCount(config.Help)
		} else {
			// the last line is an empty line but a line nonetheless
			lines = lines + 1
		}

		// since we're on one of the lines already move up one less
		for i := 0; i < lines-1; i++ {
			clearLine()
			moveUp()
		}
	}

	redraw := func() func(helpStatus) {
		var showHelp helpStatus
		return func(help helpStatus) {
			clear(showHelp)
			showHelp = help
			draw(showHelp)
		}
	}()

	draw(helpNo)

	for {
		select {
		case key := <-keyPresses:
			switch key {
			case enter:
				exitAck <- exitYes
				redraw(helpNo) // to clear help
				return config.Options[selectionIndex], Empty
			case CtrlC:
				exitAck <- exitYes
				redraw(helpNo) // to clear help
				return "", CtrlC
			case up:
				selectionIndex = ((selectionIndex - 1) + len(config.Options)) % len(config.Options)
				redraw(helpNo)
			case down:
				selectionIndex = ((selectionIndex + 1) + len(config.Options)) % len(config.Options)
				redraw(helpNo)
			case Empty:
				redraw(helpYes)
			}
		}
		exitAck <- exitNo
	}
}
