package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

type key int

const (
	unknown key = iota
	up
	down
	enter
)

func clearLine() {
	fmt.Print("\033[2K")
}

func clearScreen() {
	fmt.Print("\033[2J")
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
	//hideCursor()
	//defer showCursor()

	fmt.Printf("Use UP and DOWN arrow keys\n")

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
			default:
				keyPresses <- unknown
			}
		}
	}()

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	count := 5
	fmt.Printf("\rreturning in %d...", count)
	for {
		select {
		case <-tick.C:
			count--
			clearLine()
			fmt.Printf("\rreturning in %d...", count)
			if count == 0 {
				fmt.Println("\r")
				return
			}
		case key := <-keyPresses:
			switch key {
			case enter:
				clearLine()
				fmt.Printf("\renter")
			case up:
				clearLine()
				fmt.Printf("\rup")
			case down:
				clearLine()
				fmt.Printf("\rdown")
			case unknown:
				clearLine()
				fmt.Printf("\runknown key")
			}
		}
	}
}
