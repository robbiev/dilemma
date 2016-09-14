package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type key int

const (
	unknown key = iota
	up
	down
	enter
	ctrlc
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

func main() {
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
	go func() {
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
		}
	}()

	var lines int
	var selectionIndex int

	options := []string{"waffles", "ice cream", "candy", "biscuits"}

	draw := func() {
		fmt.Println(`Make a selection using the arrow keys:`)
		fmt.Print("\r")
		for i, v := range options {
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
		lines = len(options) + 2
	}

	clear := func() {
		// the line where we started is also filled with text so we don't need to
		// count it when moving up
		for i := 0; i < lines-1; i++ {
			clearLine()
			moveUp()
		}
	}

	draw()

	for {
		select {
		case key := <-keyPresses:
			switch key {
			case enter:
				clearLine()
				fmt.Printf("enjoy your %s\n", options[selectionIndex])
				return
			case ctrlc:
				clearLine()
				fmt.Print("exiting...\n")
				return
			case up:
				selectionIndex = ((selectionIndex - 1) + len(options)) % len(options)
				clear()
				draw()
			case down:
				selectionIndex = ((selectionIndex + 1) + len(options)) % len(options)
				clear()
				draw()
			case unknown:
				clearLine()
				fmt.Printf("Use arrow up and down, then enter to select.")
			}
		}
	}
}
