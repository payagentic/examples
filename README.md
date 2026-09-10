# PayAgentic examples

Runnable examples for developers building agent purchases and paid APIs.

[Watch the 45-second demo or download the examples](https://github.com/payagentic/examples/releases/tag/demo-v1.0.0).

## Try an API commerce exchange

**Local simulation — no account, API key, wallet or funds required.** This example runs the PayAgentic buyer SDK and merchant middleware together. Authorization and verification are simulated; it does not demonstrate live settlement.

Requires Node.js 22+ and npm. Clone the repository and run:

```sh
git clone https://github.com/payagentic/examples.git
cd examples/api-commerce
npm ci
npm test
npm start
```

Expected output:

```text
Merchant: HTTP 402; price 0.01 USDC (simulation).
Buyer: HTTP 200; received Example API data.
No funds moved. No gateway, wallet, or blockchain was contacted.
```

The tests exercise unpaid access, rejected credentials, an allowed response, policy denial and replay rejection. The merchant listens on loopback and shuts down after the run.

[Read the buyer and merchant example](api-commerce/README.md) · [Buyer source](api-commerce/buyer.mjs) · [Merchant source](api-commerce/merchant.mjs)

The example includes tested SDK packages so you can run it while public registry releases are being finalized. These are the reviewed buyer SDK 0.2.1, merchant SDK 0.1.0 and Python SDK 0.1.1 builds. The included Go source is the corresponding source snapshot.

## Explore a real integration

We are looking for two types of pilot users:

- API businesses that already sell a useful result and want to evaluate purchases by AI agents.
- Teams building agents that already purchase external services and need spending controls and transaction visibility.

Contact [hello@payagentic.ai](mailto:hello@payagentic.ai?subject=PayAgentic%20developer%20pilot) with the endpoint or workflow you want to evaluate. Pilot scope and pricing are agreed separately; this example does not create an account or incur charges.

[Developer guide](https://payagentic.ai/platform/developers?utm_source=github_examples&utm_medium=directory&utm_campaign=developer_launch) · [Public Postman workspace](https://www.postman.com/payagentic-5267497/payagentic-developer-apis/overview)

## Connect an AI application

The hosted PayAgentic MCP connector at `https://mcp.payagentic.ai/mcp` exposes six **read-only** tools for balances, policies, transactions, wallets, agents and approvals. It requires your own scoped connector credential. It does not initiate or approve payments.

Request developer access through the contact above. Never put credentials in a GitHub issue, public Postman variable or screenshot. Real payment integration and settlement are separate from this local simulation.

## License

MIT. See [LICENSE](LICENSE).
