//go:build legacy_cli

package main

import (
	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

var walletsCmd = &cobra.Command{
	Use:   "wallets",
	Short: "Manage wallets",
}

var walletsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List wallets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		opts := &payagentic.ListOptions{Limit: limit, Cursor: cursor}
		resp, err := client.Wallets.List(cmd.Context(), opts)
		if err != nil {
			return err
		}
		return printOutput(cmd, resp.Items)
	},
}

var walletsGetCmd = &cobra.Command{
	Use:   "get <wallet-id>",
	Short: "Get wallet details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		wallet, err := client.Wallets.Get(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printOutput(cmd, wallet)
	},
}

var walletsFundCmd = &cobra.Command{
	Use:   "fund <wallet-id>",
	Short: "Fund a wallet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		amount, _ := cmd.Flags().GetString("amount")
		currency, _ := cmd.Flags().GetString("currency")
		source, _ := cmd.Flags().GetString("source")
		req := &payagentic.FundWalletRequest{
			Amount:   amount,
			Currency: currency,
			Source:   source,
		}
		tx, err := client.Wallets.Fund(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		return printOutput(cmd, tx)
	},
}

var walletsWithdrawCmd = &cobra.Command{
	Use:   "withdraw <wallet-id>",
	Short: "Withdraw from a wallet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		amount, _ := cmd.Flags().GetString("amount")
		currency, _ := cmd.Flags().GetString("currency")
		destination, _ := cmd.Flags().GetString("destination")
		req := &payagentic.WithdrawWalletRequest{
			Amount:      amount,
			Currency:    currency,
			Destination: destination,
		}
		tx, err := client.Wallets.Withdraw(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		return printOutput(cmd, tx)
	},
}

var walletsFreezeCmd = &cobra.Command{
	Use:   "freeze <wallet-id>",
	Short: "Freeze a wallet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		wallet, err := client.Wallets.Freeze(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printOutput(cmd, wallet)
	},
}

func init() {
	walletsListCmd.Flags().Int("limit", 20, "Maximum number of results per page")
	walletsListCmd.Flags().String("cursor", "", "Pagination cursor")

	walletsFundCmd.Flags().String("amount", "", "Amount to fund")
	walletsFundCmd.Flags().String("currency", "USDC", "Currency")
	walletsFundCmd.Flags().String("source", "", "Funding source")
	_ = walletsFundCmd.MarkFlagRequired("amount")
	_ = walletsFundCmd.MarkFlagRequired("source")

	walletsWithdrawCmd.Flags().String("amount", "", "Amount to withdraw")
	walletsWithdrawCmd.Flags().String("currency", "USDC", "Currency")
	walletsWithdrawCmd.Flags().String("destination", "", "Withdrawal destination")
	_ = walletsWithdrawCmd.MarkFlagRequired("amount")
	_ = walletsWithdrawCmd.MarkFlagRequired("destination")

	walletsCmd.AddCommand(walletsListCmd)
	walletsCmd.AddCommand(walletsGetCmd)
	walletsCmd.AddCommand(walletsFundCmd)
	walletsCmd.AddCommand(walletsWithdrawCmd)
	walletsCmd.AddCommand(walletsFreezeCmd)
}
