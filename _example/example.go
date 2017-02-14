package main

import (
	"fmt"
	"github.com/robbiev/dilemma"
)

func main() {
	fmt.Println()

	{
		s := dilemma.Config{
			Title:      "Hello there!\n\rSelect a treat using the arrow keys:",
			Help:       "Use arrow up and down, then enter to select.\n\rChoose wisely.",
			Options:    []string{"waffles", "ice cream", "candy", "biscuits", "icy-poles", "cake", "cupcake", "muffin"},
			ShownItems: 0,
		}
		selected, exitKey, err := dilemma.Prompt(s)
		if err != nil || exitKey == dilemma.CtrlC {
			fmt.Print("Exiting...\n")
			return
		}

		fmt.Printf("Enjoy your %s!\n", selected)
	}

	fmt.Println()

	{
		s := dilemma.Config{
			Title:      "Select a companion using the arrow keys:",
			Help:       "Use arrow up and down, then enter to select.",
			Options:    []string{"dog", "pony", "cat", "rabbit", "gopher", "elephant"},
			ShownItems: 0,
		}
		selected, exitKey, err := dilemma.Prompt(s)
		if err != nil || exitKey == dilemma.CtrlC {
			fmt.Print("Exiting...\n")
			return
		}

		fmt.Printf("Enjoy your %s!\n", selected)
	}

	fmt.Println()
}
