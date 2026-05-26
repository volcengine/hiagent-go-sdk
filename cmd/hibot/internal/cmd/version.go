package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version is the CLI version (overridable via -ldflags at build time).
var (
	Version   = "0.1.0"
	Commit    = "none"
	BuildDate = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the hibot CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "hibot %s %s/%s (commit %s, built %s)\n",
				Version, runtime.GOOS, runtime.GOARCH, Commit, BuildDate)
			return nil
		},
	}
}
