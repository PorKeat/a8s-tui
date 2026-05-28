package main

import (
	"fmt"
	"os"

	"github.com/ITProfessional-Gen01/a8s-cli/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running a8s-cli: %v\n", err)
		os.Exit(1)
	}
}
