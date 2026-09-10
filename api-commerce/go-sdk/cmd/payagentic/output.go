//go:build legacy_cli

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	payagentic "github.com/payagentic/payagentic-go"
	"github.com/spf13/cobra"
)

// printOutput renders data in the format specified by the --output flag.
func printOutput(cmd *cobra.Command, data any) error {
	format := getOutputFormat(cmd)
	switch format {
	case "json":
		return printJSON(data)
	default:
		return printTable(data)
	}
}

// printJSON marshals data to indented JSON and writes it to stdout.
func printJSON(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// printTable renders data as an aligned table to stdout.
func printTable(data any) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	switch v := data.(type) {
	case []payagentic.Wallet:
		fmt.Fprintln(w, "ID\tADDRESS\tCHAIN\tBALANCE\tSTATUS")
		for _, item := range v {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", item.ID, item.Address, item.Chain, item.Balance, item.Status)
		}
	case *payagentic.Wallet:
		fmt.Fprintln(w, "ID\tADDRESS\tCHAIN\tBALANCE\tSTATUS")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", v.ID, v.Address, v.Chain, v.Balance, v.Status)
	case []payagentic.Agent:
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED")
		for _, item := range v {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.ID, item.Name, item.Status, item.CreatedAt.Format("2006-01-02"))
		}
	case *payagentic.Agent:
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Status, v.CreatedAt.Format("2006-01-02"))
	case []payagentic.Transaction:
		fmt.Fprintln(w, "ID\tWALLET\tTYPE\tAMOUNT\tSTATUS\tCREATED")
		for _, item := range v {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s\t%s\n", item.ID, item.WalletID, item.Type, item.Amount, item.Currency, item.Status, item.CreatedAt.Format("2006-01-02"))
		}
	case *payagentic.Transaction:
		fmt.Fprintln(w, "ID\tWALLET\tTYPE\tAMOUNT\tSTATUS\tCREATED")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s\t%s\n", v.ID, v.WalletID, v.Type, v.Amount, v.Currency, v.Status, v.CreatedAt.Format("2006-01-02"))
	case []payagentic.Policy:
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tCREATED")
		for _, item := range v {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.ID, item.Name, item.Type, item.CreatedAt.Format("2006-01-02"))
		}
	case *payagentic.Policy:
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tCREATED")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Type, v.CreatedAt.Format("2006-01-02"))
	case []payagentic.Approval:
		fmt.Fprintln(w, "ID\tTYPE\tSTATUS\tREQUESTED BY\tCREATED")
		for _, item := range v {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", item.ID, item.Type, item.Status, item.RequestedBy, item.CreatedAt.Format("2006-01-02"))
		}
	case *payagentic.Approval:
		fmt.Fprintln(w, "ID\tTYPE\tSTATUS\tREQUESTED BY\tCREATED")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", v.ID, v.Type, v.Status, v.RequestedBy, v.CreatedAt.Format("2006-01-02"))
	case *payagentic.RotateKeyResponse:
		fmt.Fprintln(w, "AGENT ID\tKEY FINGERPRINT\tAPI KEY")
		fmt.Fprintf(w, "%s\t%s\t%s\n", v.AgentID, v.KeyFingerprint, v.APIKey)
	default:
		// Fall back to JSON for unknown types.
		return printJSON(data)
	}
	return nil
}
