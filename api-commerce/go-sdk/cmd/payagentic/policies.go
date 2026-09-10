//go:build legacy_cli

package main

import (
	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

var policiesCmd = &cobra.Command{
	Use:   "policies",
	Short: "Manage spend policies",
}

var policiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		opts := &payagentic.ListOptions{Limit: limit, Cursor: cursor}
		resp, err := client.Policies.List(cmd.Context(), opts)
		if err != nil {
			return err
		}
		return printOutput(cmd, resp.Items)
	},
}

var policiesGetCmd = &cobra.Command{
	Use:   "get <policy-id>",
	Short: "Get policy details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		policy, err := client.Policies.Get(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printOutput(cmd, policy)
	},
}

var policiesApplyCmd = &cobra.Command{
	Use:   "apply <policy-id> <wallet-id>",
	Short: "Apply a policy to a wallet (placeholder)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Placeholder: policy application is typically done via the API's wallet
		// or policy endpoints. This will be wired up once the backend endpoint exists.
		cmd.Printf("Policy %s applied to wallet %s (placeholder).\n", args[0], args[1])
		return nil
	},
}

func init() {
	policiesListCmd.Flags().Int("limit", 20, "Maximum number of results per page")
	policiesListCmd.Flags().String("cursor", "", "Pagination cursor")

	policiesCmd.AddCommand(policiesListCmd)
	policiesCmd.AddCommand(policiesGetCmd)
	policiesCmd.AddCommand(policiesApplyCmd)
}
