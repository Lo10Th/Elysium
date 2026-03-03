package main

import (
	"os"

	"github.com/elysium/elysium/cli/cmd"
)

func main() {
	cmd.Execute()
	os.Exit(0)
}
