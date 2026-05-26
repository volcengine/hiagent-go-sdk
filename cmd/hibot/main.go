package main

import (
	"fmt"
	"os"

	"github.com/volcengine/hiagent-go-sdk/cmd/hibot/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cmd.ExitCodeFor(err))
	}
}
