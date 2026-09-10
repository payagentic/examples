# Connect to PayAgentic over MCP

Use the hosted connector to inspect your PayAgentic account from an MCP-compatible AI application. The connector is read-only: it does not initiate or approve payments.

## Connection settings

In a client that supports remote Streamable HTTP MCP servers and custom headers, add:

| Setting | Value |
| --- | --- |
| URL | `https://mcp.payagentic.ai/mcp` |
| Transport | Streamable HTTP |
| Authorization header | `Authorization: Bearer <your scoped connector credential>` |
| Optional header | `X-Organization-Id`, only if required by your credential |

Request developer access at [hello@payagentic.ai](mailto:hello@payagentic.ai?subject=PayAgentic%20MCP%20access). Store the credential in the client's private credential field or local secret store. Do not put it in a public repository or issue. The current connector uses a bearer API key; it does not provide an OAuth login flow.

## Available tools

| Tool | Purpose |
| --- | --- |
| `get_balance` | Read a wallet balance |
| `get_policy_for_intent` | Read the policy for a payment intent |
| `list_recent_transactions` | Inspect recent transaction records |
| `list_wallets` | List accessible wallets |
| `payagentic_list_agents` | List accessible agents |
| `payagentic_list_approvals` | Inspect approvals |

After connecting, confirm that your client lists these six tools. Start with `list_wallets`; an empty list can be a valid result for a new account. HTTP 401 means a missing or invalid credential. An authorization failure is not fixed by changing the public hostname.

The [public server card](https://mcp.payagentic.ai/.well-known/mcp/server-card.json) describes tool schemas without exposing account data. Authenticated initialization, discovery, a read-only wallet call and subsequent test-key revocation were verified on September 9, 2026. Client and directory installation flows require their own verification.

## Discovery metadata

[server.json](server.json) is the prepared Official MCP Registry manifest. It is listing metadata, not a client configuration file and not evidence of a published registry entry. Namespace ownership verification and directory publication are still being finalized.

## Try a local example first

The [buyer-and-merchant demo](../api-commerce/README.md) requires no account or funds. It simulates authorization and verification and is separate from the authenticated MCP connection and live settlement.
