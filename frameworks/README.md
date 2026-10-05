# Read-only framework examples

Runnable examples for LangChain, LangGraph, Vercel AI SDK and CrewAI using the published PayAgentic SDKs. Each exposes one native tool: `payagentic_wallet_page_summary`. It counts wallets on the **first page**, and returns `has_more`; it does not claim to count all wallets in an account.

The demos use synthetic fixtures. No PayAgentic account, model-provider key, payment or external API call is needed after dependency installation. They demonstrate framework execution and SDK compatibility, not live service availability, OAuth acceptance or marketplace approval.

## JavaScript: LangChain, LangGraph and Vercel AI SDK

Requires Node.js 22 or newer.

```sh
cd frameworks/javascript
npm ci --ignore-scripts --no-audit --no-fund
npm test
npm start
```

`tools.mjs` creates a native LangChain structured tool and a Vercel `tool` with `inputSchema` and `execute`. `demo.mjs` also invokes the LangChain tool through LangGraph's `ToolNode`. Tests exercise Vercel's `generateText` dispatch with its local mock model; no paid model is called.

For application integration, pass an initialized `PayagenticClient` to `createLangChainTool(client)` or `createVercelTools(client)`. The returned objects can be supplied to the respective framework's tool collection. Keep the SDK client on the server and load its credential from your secret manager. Use an isolated account with only the access needed for wallet reads. Do not paste credentials into source, prompts, screenshots, issues or tool inputs.

## Python: CrewAI

Requires Python 3.13 and [uv](https://docs.astral.sh/uv/). The pinned PayAgentic release requires Python >=3.13, while the pinned CrewAI release requires Python <3.14.

```sh
cd frameworks/crewai
uv sync --locked
uv run --no-sync python -m unittest -v test_tool
uv run --no-sync python demo.py
```

`WalletPageSummary(client)` is a real CrewAI `BaseTool` with a validated input schema and `_run` implementation, backed by the released Python SDK. Pass it in your CrewAI agent's `tools` list. The fixture uses a temporary HTTP server bound to `127.0.0.1`, with synthetic data. It never talks to the production service.

The demo and tests disable tracing and telemetry, use temporary storage and simulate a logged-out CrewAI cloud user. This avoids CrewAI 1.15.23's import-time access to its saved cloud login. That isolation helper is only for the fixture; application code should import `wallet_tool.py` directly and manage its own framework configuration. The wrapper uses the SDK's generated `PaginatedWallets` model, so revalidate it when upgrading the pinned SDK.

Known SDK limitation: PayAgentic Python 0.1.1 can leave an HTTP connection open after a failed request, producing a `ResourceWarning` in the negative tests. A transport cleanup fix is awaiting release. These fixtures are suitable for trying the integration locally; upgrade and revalidate the SDK before adopting this pattern in a long-running service.

## Data handling and limits

- Only `GET /v1/wallets` is wired into these tools. No payment or approval tool is registered.
- Outputs contain only `wallets_on_page` and `has_more`. Wallet identifiers, addresses, labels, balances, response headers and credentials are not forwarded to the model.
- The SDK still receives the wallet response in application memory. Do not enable raw HTTP logging or account-data tracing in a live application. Even the aggregate count is account information; use an approved model and appropriate access controls.
- Unknown fields, non-integer limits and limits outside 1–100 are rejected before calling the SDK. Upstream failures produce a fixed message rather than raw response text.
- A local fixture pass does not prove that a real account has access. Live acceptance requires a disposable, scoped account and a healthy service. No live acceptance is claimed here.

Pinned dependencies and lockfiles make the examples reproducible. Tests run in GitHub Actions without repository secrets. Re-run them and review dependency changes before upgrading.

## Framework references

- [LangChain tools](https://docs.langchain.com/oss/javascript/langchain/tools)
- [LangGraph ToolNode](https://reference.langchain.com/javascript/langchain-langgraph/prebuilt/ToolNode)
- [Vercel AI SDK tools](https://ai-sdk.dev/docs/foundations/tools)
- [CrewAI custom tools](https://docs.crewai.com/en/learn/create-custom-tools)
