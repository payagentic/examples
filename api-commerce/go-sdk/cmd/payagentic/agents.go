//go:build legacy_cli

package main

import (
	"fmt"

	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage agents",
}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		opts := &payagentic.ListOptions{Limit: limit, Cursor: cursor}
		resp, err := client.Agents.List(cmd.Context(), opts)
		if err != nil {
			return err
		}
		return printOutput(cmd, resp.Items)
	},
}

var agentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		name, _ := cmd.Flags().GetString("name")
		req := &payagentic.CreateAgentRequest{Name: name}
		agent, err := client.Agents.Create(cmd.Context(), req)
		if err != nil {
			return err
		}
		return printOutput(cmd, agent)
	},
}

var agentsRotateKeyCmd = &cobra.Command{
	Use:   "rotate-key <agent-id>",
	Short: "Rotate an agent's API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		resp, err := client.Agents.RotateKey(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printOutput(cmd, resp)
	},
}

var agentsRevokeCmd = &cobra.Command{
	Use:   "revoke <agent-id>",
	Short: "Revoke (delete) an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient(cmd)
		if err := client.Agents.Delete(cmd.Context(), args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Agent %s revoked successfully.\n", args[0])
		return nil
	},
}

func init() {
	agentsListCmd.Flags().Int("limit", 20, "Maximum number of results per page")
	agentsListCmd.Flags().String("cursor", "", "Pagination cursor")

	agentsCreateCmd.Flags().String("name", "", "Agent name")
	_ = agentsCreateCmd.MarkFlagRequired("name")

	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsCreateCmd)
	agentsCmd.AddCommand(agentsRotateKeyCmd)
	agentsCmd.AddCommand(agentsRevokeCmd)
}
