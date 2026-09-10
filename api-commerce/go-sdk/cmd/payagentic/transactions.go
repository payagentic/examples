//go:build legacy_cli

package main

import (
	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

var transactionsCmd = &cobra.Command{
	Use:     "transactions",
	Aliases: []string{"tx"},
	Short:   "View transactions",
}

var transactionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		opts := &payagentic.ListOptions{Limit: limit, Cursor: cursor}
		resp, err := client.Transactions.List(cmd.Context(), opts)
		if err != nil {
			return err
		}
		return printOutput(cmd, resp.Items)
	},
}

var transactionsGetCmd = &cobra.Command{
	Use:   "get <transaction-id>",
	Short: "Get transaction details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		tx, err := client.Transactions.Get(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printOutput(cmd, tx)
	},
}

func init() {
	transactionsListCmd.Flags().Int("limit", 20, "Maximum number of results per page")
	transactionsListCmd.Flags().String("cursor", "", "Pagination cursor")

	transactionsCmd.AddCommand(transactionsListCmd)
	transactionsCmd.AddCommand(transactionsGetCmd)
}
