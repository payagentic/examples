# An agent buys API data; a merchant serves it

Run both sides of a paid API exchange using the PayAgentic buyer SDK and Express merchant middleware. This local simulation needs **Node.js 22+ and npm**, with internet access for dependency installation. It needs no account, API key, wallet or funds.

From the repository root, open `api-commerce` and run:

```sh
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

`merchant.mjs` protects `/data` with `payagentic({ ... })` middleware. `buyer.mjs` uses `PaymentsClient.fetch()` to read the price, obtain a simulated credential and retry. The merchant serves data only after the fixture verifier accepts the credential. The server listens on loopback and closes after the run. Tests cover unpaid access, invalid credentials, successful access, policy denial and replay rejection.

The bundled SDK tarballs and Python wheel are built from the same source revision as this example. The merchant SDK is not yet on the public npm registry. Do not replace its local dependency with an unavailable registry install command. The buyer package is also available as [@payagentic.ai/sdk on npm](https://www.npmjs.com/package/@payagentic.ai/sdk); published versions may lag these source builds.

## Python connection setup

The ZIP also includes a source-built Python wheel. The currently published npm 0.2.0 package fails a plain Node import; PyPI 0.1.0 omits a required dependency. Use these tested bundled builds until updated releases are published.

```sh
python3.13 -m venv .venv
. .venv/bin/activate
python -m pip install ./payagentic-*.whl
```

Verify the installed package without an account:

```sh
python -c "from payagentic import PayAgentic, AsyncPayAgentic, __version__; print(__version__)"
```

For an authenticated Python integration, follow the [developer guide](https://payagentic.ai/platform/developers) or contact hello@payagentic.ai. The local buyer and merchant demo uses Node.js.

## Go connection setup

The `go-sdk` directory contains the matching source snapshot and requires Go 1.25+. The bundled `go-sdk/examples/connection` provides a read-only connection check. From `api-commerce`, run:

```sh
cd go-sdk
go run ./examples/connection
```

Supply `PAYAGENTIC_API_KEY` and `PAYAGENTIC_BASE_URL` in the runtime environment. The source snapshot includes newer APIs than the public Go release mirror, so use this replacement when trying the website's current payment examples.

## Connect a real test integration

This demo deliberately simulates authorization and verification. It does **not** demonstrate settlement, fees, origin verification or on-chain transfers. Never expose the fixture verifier publicly or use it to authorize real work.

1. Follow the [quickstart](https://payagentic.ai/docs/quickstart) to provision an agent, finish wallet provisioning and configure spend policies. Supply the assigned test API key, wallet ID and gateway URL through your runtime environment.
2. As a merchant, register your API and priced endpoint, prove ownership of its public HTTPS origin, and configure the vendor SDK's `sellerBinding` with the registered IDs, test network, asset and a wallet-provider signing callback. See the vendor README included in `payagentic-vendor.tgz`. A direct unregistered payment does not demonstrate PayAgentic's registered merchant fee flow.
3. Replace the fixture proposer with the gateway-backed client and replace the fixture verifier with the gateway. Use a funded test wallet and a merchant in the same test environment. Keep amounts as decimal strings.
4. Confirm allowed, denied and approval-required outcomes. Reconcile the gateway transaction status, merchant delivery and fee ledger before considering live use. A successful HTTP response alone is not settlement evidence.

The buyer SDK includes `examples/buy-api.mjs` for a configured test gateway. It requires explicit wallet, URL, spend limit and idempotency settings and never prints keys or paid content.

## Troubleshooting

- Missing `.tgz` files: clone this repository again or download its release archive; the tarballs are included in `api-commerce`.
- Missing imports or engine errors: use Node 22+ and run `npm install` in the extracted folder.
- A denied proposal is expected to stop the buyer. Investigate the decision before retrying; do not switch to a live key to bypass it.
- Real gateway failures: confirm the configured gateway with your deployment operator. Current SDK defaults may differ from your deployed ingress. The example's `.invalid` URL is intentionally never contacted.

[SDK documentation](https://payagentic.ai/sdks) · [Pricing](https://payagentic.ai/pricing) · [Integration support](mailto:hello@payagentic.ai)
