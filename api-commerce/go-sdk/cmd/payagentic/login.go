//go:build legacy_cli

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via OIDC device flow",
	Long: `Authenticate with PayAgentic using the OIDC device authorization flow.

This initiates a device code flow where you will be given a URL and code
to enter in your browser. Once authenticated, your credentials are stored
locally for subsequent CLI commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Placeholder: OIDC device flow implementation will follow once the
		// identity service OIDC endpoints are available.
		fmt.Fprintln(cmd.OutOrStdout(), "OIDC device flow login is not yet implemented.")
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "In the meantime, authenticate using an API key:")
		fmt.Fprintln(cmd.OutOrStdout(), "  export PAYAGENTIC_API_KEY=your-api-key")
		fmt.Fprintln(cmd.OutOrStdout(), "  payagentic wallets list")
		return nil
	},
}
