package main

import (
	"os"

	"github.com/monopole/gojira/internal/commands"
)

func main() {
	if err := commands.NewGoJiraCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
