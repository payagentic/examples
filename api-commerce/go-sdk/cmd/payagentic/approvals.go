//go:build legacy_cli

package main

import (
	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

var approvalsCmd = &cobra.Command{
	Use:   "approvals",
	Short: "Manage approval requests",
}

var approvalsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending approvals",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		opts := &payagentic.ListOptions{Limit: limit, Cursor: cursor}
		resp, err := client.Approvals.List(cmd.Context(), opts)
		if err != nil {
			return err
		}
		return printOutput(cmd, resp.Items)
	},
}

var approvalsApproveCmd = &cobra.Command{
	Use:   "approve <approval-id>",
	Short: "Approve a pending request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		reason, _ := cmd.Flags().GetString("reason")
		req := &payagentic.ApprovalDecisionRequest{Reason: reason}
		approval, err := client.Approvals.Approve(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		return printOutput(cmd, approval)
	},
}

var approvalsDenyCmd = &cobra.Command{
	Use:   "deny <approval-id>",
	Short: "Deny a pending request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		reason, _ := cmd.Flags().GetString("reason")
		req := &payagentic.ApprovalDecisionRequest{Reason: reason}
		approval, err := client.Approvals.Deny(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		return printOutput(cmd, approval)
	},
}

func init() {
	approvalsListCmd.Flags().Int("limit", 20, "Maximum number of results per page")
	approvalsListCmd.Flags().String("cursor", "", "Pagination cursor")

	approvalsApproveCmd.Flags().String("reason", "", "Reason for approval")
	approvalsDenyCmd.Flags().String("reason", "", "Reason for denial")

	approvalsCmd.AddCommand(approvalsListCmd)
	approvalsCmd.AddCommand(approvalsApproveCmd)
	approvalsCmd.AddCommand(approvalsDenyCmd)
}
