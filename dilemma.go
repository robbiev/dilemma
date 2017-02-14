// Package dilemma allows creating a TTY selection prompt.
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

type input struct {
	key Key
	err error
}

type exitStatus int

type helpStatus int

// Config holds the configuration to display a list of options
// for a user to select.
type Config struct {
	Title      string
	Options    []string
	Help       string
	ShownItems int
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

func inputLoop(keyPresses chan<- input, exitAck chan exitStatus) {
	buf := make([]byte, 128)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			keyPresses <- input{key: Empty, err: err}
			return
		}
		bufstr := string(buf[:n])
		switch {
		case bufstr == "\033[A":
			keyPresses <- input{key: up}
		case bufstr == "\033[B":
			keyPresses <- input{key: down}
		case bufstr == "\x0D":
			keyPresses <- input{key: enter}
		case bufstr == "\x03":
			keyPresses <- input{key: CtrlC}
		default:
			keyPresses <- input{key: Empty}
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
func Prompt(config Config) (string, Key, error) {
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		return "", Empty, err
	}
	defer terminal.Restore(0, oldState)

	hideCursor()
	defer showCursor()

	// ensure we always exit with the cursor at the beginning of the line so the
	// terminal prompt prints in the expected place
	defer func() {
		fmt.Print("\r")
	}()

	keyPresses := make(chan input, 1)
	exitAck := make(chan exitStatus)
	go inputLoop(keyPresses, exitAck)

	var selectionIndex int

	draw := func(help helpStatus) {
		fmt.Println(config.Title)
		fmt.Print("\r")
		for i, v := range config.Options {
			minVal := 0
			maxVal := 0
			if selectionIndex < config.ShownItems {
				minVal = 0
			} else {
				if config.ShownItems != 0 {
					minVal = selectionIndex - config.ShownItems
				} else {
					minVal = 0
				}
			}
			if selectionIndex+config.ShownItems <= len(config.Options) {
				maxVal = selectionIndex + config.ShownItems
			} else {
				maxVal = len(config.Options)
			}
			if (i >= minVal && i <= maxVal) || (config.ShownItems == 0) {
				if i == selectionIndex {
					invertColours()
				}
				fmt.Printf("%s\n", v)
				if i == selectionIndex {
					resetStyle()
				}
				fmt.Print("\r")
			}
		}
		if (config.ShownItems < len(config.Options)) || config.ShownItems != 0 {
			fmt.Print("...")
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
		input := <-keyPresses
		if input.err != nil {
			redraw(helpNo) // to clear help
			return "", Empty, input.err
		}
		switch input.key {
		case enter:
			exitAck <- exitYes
			redraw(helpNo) // to clear help
			return config.Options[selectionIndex], Empty, nil
		case CtrlC:
			exitAck <- exitYes
			redraw(helpNo) // to clear help
			return "", CtrlC, nil
		case up:
			selectionIndex = ((selectionIndex - 1) + len(config.Options)) % len(config.Options)
			redraw(helpNo)
		case down:
			selectionIndex = ((selectionIndex + 1) + len(config.Options)) % len(config.Options)
			redraw(helpNo)
		case Empty:
			redraw(helpYes)
		}
		exitAck <- exitNo
	}
}
