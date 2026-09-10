//go:build legacy_cli

package main

import (
	"fmt"
	"runtime"

	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "payagentic CLI v%s\n", payagentic.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "  SDK:      payagentic-go v%s\n", payagentic.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "  Go:       %s\n", runtime.Version())
		fmt.Fprintf(cmd.OutOrStdout(), "  OS/Arch:  %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
