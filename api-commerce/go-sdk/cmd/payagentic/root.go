//go:build legacy_cli

package main

import (
	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "payagentic",
	Short: "PayAgentic CLI for managing agent wallets, payments, and policies",
	Long: `payagentic is the command-line interface for the PayAgentic agent payments platform.

Manage programmable USDC wallets, spend policies, and x402 payments
for AI agents operating across African markets.`,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication (env: PAYAGENTIC_API_KEY)")
	rootCmd.PersistentFlags().String("base-url", "", "API base URL (env: PAYAGENTIC_BASE_URL)")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table or json")

	rootCmd.AddCommand(walletsCmd)
	rootCmd.AddCommand(agentsCmd)
	rootCmd.AddCommand(transactionsCmd)
	rootCmd.AddCommand(policiesCmd)
	rootCmd.AddCommand(approvalsCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(versionCmd)
}

// getClient builds a PayAgentic client from CLI flags and environment variables.
func getClient(cmd *cobra.Command) *payagentic.Client {
	var opts []payagentic.Option

	// Environment variables are the baseline.
	opts = append(opts, payagentic.FromEnvironment())

	// CLI flags override environment variables.
	if key, _ := cmd.Flags().GetString("api-key"); key != "" {
		opts = append(opts, payagentic.WithAPIKey(key))
	}
	if baseURL, _ := cmd.Flags().GetString("base-url"); baseURL != "" {
		opts = append(opts, payagentic.WithBaseURL(baseURL))
	}

	return payagentic.NewClient(opts...)
}

// getOutputFormat returns the requested output format (table or json).
func getOutputFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("output")
	return format
}
