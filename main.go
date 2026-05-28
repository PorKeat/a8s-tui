package main

import (
	"fmt"
	"os"

	"github.com/PorKeat/a8s-tui/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running a8s-tui: %v\n", err)
		os.Exit(1)
	}
}
