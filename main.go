package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type key int

type exitStatus int

type interruptStatus int

type helpStatus int

type selection struct {
	title   string
	options []string
	help    string
}

const (
	unknown key = iota
	up
	down
	enter
	ctrlc
)

const (
	exitNo exitStatus = iota
	exitYes
)

const (
	intNo interruptStatus = iota
	intYes
)

const (
	helpNo helpStatus = iota
	helpYes
)

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

func inputLoop(keyPresses chan<- key, exitAck chan exitStatus) {
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
			keyPresses <- ctrlc
		default:
			keyPresses <- unknown
		}
		if exitYes == <-exitAck {
			return
		}
	}
}

func promptSelection(sel selection) (string, interruptStatus) {
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

	keyPresses := make(chan key)
	exitAck := make(chan exitStatus)
	go inputLoop(keyPresses, exitAck)

	var selectionIndex int

	draw := func(help helpStatus) {
		fmt.Println(sel.title)
		fmt.Print("\r")
		for i, v := range sel.options {
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
			fmt.Print(sel.help)
		}
	}

	clear := func(help helpStatus) {
		lines := lineCount(sel.title) + len(sel.options)

		if help == helpYes {
			lines = lines + lineCount(sel.help)
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
				return sel.options[selectionIndex], intNo
			case ctrlc:
				exitAck <- exitYes
				redraw(helpNo) // to clear help
				return "", intYes
			case up:
				selectionIndex = ((selectionIndex - 1) + len(sel.options)) % len(sel.options)
				redraw(helpNo)
			case down:
				selectionIndex = ((selectionIndex + 1) + len(sel.options)) % len(sel.options)
				redraw(helpNo)
			case unknown:
				redraw(helpYes)
			}
		}
		exitAck <- exitNo
	}
}

func main() {
	fmt.Println()

	{
		s := selection{
			title:   "Hello friend!\n\rSelect a treat using the arrow keys:",
			help:    "Use arrow up and down, then enter to select.\n\rChoose wisely.",
			options: []string{"waffles", "ice cream", "candy", "biscuits"},
		}
		selected, interrupt := promptSelection(s)
		if interrupt == intYes {
			fmt.Print("Exiting...\n")
			return
		}

		fmt.Printf("Enjoy your %s!\n", selected)
	}

	fmt.Println()

	{
		s := selection{
			title:   "Select a companion using the arrow keys:",
			help:    "Use arrow up and down, then enter to select.",
			options: []string{"dog", "pony", "cat", "rabbit", "gopher", "elephant"},
		}
		selected, interrupt := promptSelection(s)
		if interrupt == intYes {
			fmt.Print("Exiting...\n")
			return
		}

		fmt.Printf("Enjoy your %s!\n", selected)
	}

	fmt.Println()
}
