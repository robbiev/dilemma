package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"
)

type key int

type cursorPos struct {
	row, col int
}

const (
	unknown key = iota
	up
	down
	enter
)

var (
	cursorPosRegex = regexp.MustCompile("^\033\\[([0-9]+);([0-9]+)R$")
)

func clearLine() {
	fmt.Print("\033[2K")
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func fullScreenEnter() {
	fmt.Printf("\033[?1049h\033[H")
}

func fullScreenExit() {
	fmt.Printf("\033[?1049l")
}

func main() {
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(0, oldState)
	hideCursor()
	defer showCursor()

	keyPresses := make(chan key)
	cursorPosReply := make(chan cursorPos)
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
			case cursorPosRegex.MatchString(input):
				matches := cursorPosRegex.FindStringSubmatch(input)
				row, col := matches[1], matches[2]
				//fmt.Println("row", row)
				//fmt.Println("col", col)
				r, _ := strconv.Atoi(row)
				c, _ := strconv.Atoi(col)
				cursorPosReply <- cursorPos{row: r, col: c}
			default:
				keyPresses <- unknown
			}
		}
	}()

	// ask for the cursor position
	fmt.Printf("\033[6n")

	pos := <-cursorPosReply

	_, height, _ := terminal.GetSize(0)

	var lines int
	var selection int

	options := []string{"waffles", "ice cream", "candy", "biscuits"}

	draw := func() {
		selectionIndex := int(math.Abs(float64(selection % len(options))))

		fmt.Println(`Make a selection using the arrow keys:`)
		fmt.Print("\r")
		for i, v := range options {
			fmt.Print("  ")
			if i == selectionIndex {
				fmt.Print("\033[7m")
			}
			fmt.Printf("%s\n", v)
			if i == selectionIndex {
				fmt.Print("\033[0m")
			}
			fmt.Print("\r")
		}
		lines = len(options) + 2
	}

	clear := func() {
		// the line where we started is also filled with text so we don't need to
		// count it when moving up
		moveOffset := lines - 1
		// correct the position when we're at the bottom of the screen
		correct := height - pos.row
		correct = moveOffset - int(math.Min(float64(correct), float64(moveOffset)))

		// set the cursor to where we started
		fmt.Printf("\033[%d;%dH", pos.row-correct, pos.col)

		// erase from the cursor onwards
		fmt.Printf("\033[J")
	}

	draw()

	for {
		select {
		case key := <-keyPresses:
			switch key {
			case enter:
				clearLine()
				selectionIndex := int(math.Abs(float64(selection % len(options))))
				fmt.Printf("\renjoy your %s\n\r", options[selectionIndex])
				return
			case up:
				selection--
				clear()
				draw()
			case down:
				selection++
				clear()
				draw()
			case unknown:
				clearLine()
				fmt.Printf("\ruse arrow up and down, then enter to select")
			}
		}
	}
}
