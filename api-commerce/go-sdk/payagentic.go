// Package payagentic provides a Go client for the PayAgentic agent payments platform.
//
// PayAgentic enables African businesses building AI agents to manage programmable
// USDC wallets. Agents pay external APIs over x402, governed by operator-defined
// spend policies. Operators fund wallets via local fiat rails and withdraw to
// local bank accounts.
//
// # Quick start
//
//	client := payagentic.NewClient(
//	    payagentic.WithAPIKey("your-api-key"),
//	)
//
//	wallets, err := client.Wallets.List(ctx, nil)
//
// # Configuration from environment
//
//	client := payagentic.NewClient(payagentic.FromEnvironment())
//
// Set PAYAGENTIC_API_KEY and PAYAGENTIC_AGENT_ID environment variables.
package payagentic

// Version is the current SDK version.
const Version = "0.1.0"
