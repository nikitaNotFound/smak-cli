package main

import (
	"os"

	"github.com/nikitaNotFound/smak-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
