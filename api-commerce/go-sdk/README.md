# payagentic-go

Build AI agents that buy paid API data with programmable stablecoin wallets and spend-policy controls on supported USDC and USDT rails. PayAgentic SDKs connect your server to the gateway; the platform's subscriptions and transaction charges are separate from the SDK's MIT license.

[Developer home](https://payagentic.ai/platform/developers) · [Runnable buyer + merchant example](https://payagentic.ai/docs/examples) · [SDK reference](https://payagentic.ai/sdks) · [Pricing](https://payagentic.ai/pricing) · [Integration support](mailto:developers@payagentic.ai)

## Start with a working example

Download the [local API commerce demo](https://payagentic.ai/examples/api-commerce.zip). It includes both sides, installation instructions, expected output and automated tests. Node.js 22+ runs the demo, including for Python and Go developers who want to inspect the HTTP payment exchange. It uses simulated authorization, moves no funds, and does not validate settlement.

For your own integration, use the [quickstart](https://payagentic.ai/docs/quickstart), finish asynchronous wallet provisioning, and configure a test API key, gateway URL and wallet ID. The key belongs in your server environment, never browser code. Registered merchant origin and endpoint setup is required to exercise the platform transaction fee flow.

These are pre-1.0 SDKs. A source checkout may be newer than the published package; pin and test a release before upgrading. See the registry's release history for available versions.

## Install

```bash
go get github.com/payagentic/payagentic-go
```

Requires Go 1.25+ (see `go.mod`).

> This public repository is a release mirror. SDK development happens in the private PayAgentic monorepo; use this repository to install releases and inspect public source. Report integration issues through [support](mailto:developers@payagentic.ai).

## Usage

### Check your connection (read-only)

Save this as `main.go`, provide `PAYAGENTIC_API_KEY` and `PAYAGENTIC_BASE_URL` through your runtime environment, then run `go run .`. It prints the HTTP status without exposing wallet data or credentials.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    payagentic "github.com/payagentic/payagentic-go"
)

func main() {
    baseURL := os.Getenv("PAYAGENTIC_BASE_URL")
    if baseURL == "" { log.Fatal("Set PAYAGENTIC_BASE_URL") }
    client, err := payagentic.NewClient(payagentic.WithBaseURL(baseURL))
    if err != nil { log.Fatal("Check SDK configuration") }
    defer client.Close()
    result, err := client.OpenAPI.ListWalletsWithResponse(context.Background(), nil)
    if err != nil { log.Fatal("Gateway request failed") }
    fmt.Println("Gateway HTTP status:", result.StatusCode())
}
```

The same program is in `examples/connection/main.go`; from a source checkout run `go run ./examples/connection`.

For paid requests also configure `PAYAGENTIC_WALLET_ID` or `WithWalletID(...)`, a funded test wallet, spend policies and your test merchant endpoint. Start with the [integration setup](https://payagentic.ai/docs/examples).

The remaining snippets illustrate individual methods inside an application with an initialized client and context.

### Invoke a public capability

Capability invocation takes an explicit `payagentic.CapabilityInput` JSON
object and a caller-owned idempotency key:

```go
result, err := client.Capabilities.Invoke(
    ctx,
    "acme",
    "answer-invoices",
    payagentic.CapabilityInput{
        "question": "What is the payment window?",
        "context": []any{"invoice-123", true},
    },
    payagentic.InvokeCapabilityOptions{IdempotencyKey: "one-logical-run"},
)
```

Nested values may be `nil`, `bool`, valid UTF-8 `string`, finite `float64`,
native integers, `json.Number`, `[]any`, or `map[string]any`. Native integers
and `json.Number` values are accepted only when their exact mathematical value
is a finite IEEE-754 binary64 value; `json.Number` lexemes are capped at 2 KiB
before exact validation. Use a JSON string for larger or higher-precision
numbers. A `float64` is already a binary64 value and is serialized with RFC
8785's shortest ECMAScript spelling.

The capability boundary copies maps and slices, rejects invalid UTF-8 before
Go can replace it, and rejects structs, typed containers, `json.RawMessage`,
and custom JSON/text marshalers. The final canonical request must be at most 64
KiB. This explicit domain keeps transport bytes consistent with TypeScript and
Python without trying to reproduce `encoding/json` field-selection rules.

That restricted domain is specific to capability invocation. `SignJWS` and
the exported `CanonicalJSONBytes` helper accept ordinary values supported by
`encoding/json`, including structs and tags, typed maps and slices,
`json.RawMessage`, and custom JSON/text marshalers. `CanonicalJSONBytes`
marshals once, validates the emitted UTF-8/I-JSON bytes, and rejects integral
claims that would change under RFC 8785's binary64 conversion before applying
the pinned JCS transform. Emitted numeric tokens are capped at 2 KiB before
exact validation; this covers every finite binary64 JCS spelling and full
exact binary64 decimal expansions, including minimum subnormal. Shortest RFC
8785 nonintegral spellings such as `0.1` remain valid. Encode larger or
higher-precision claims as JSON strings. This does not apply capability input's
stricter exact-source number rules.

### x402 paywall walker

```go
resp, err := client.X402.Fetch(ctx, "https://api.vendor.com/paid-endpoint")
```

Hit a paid URL, the walker proposes a payment from your wallet, settles it via the gateway, then retries with `X-Payment-Receipt`.

### Typed errors

Non-2xx responses become typed Go errors. Use `errors.As` to discriminate:

```go
import (
    "errors"
    payagentic "github.com/payagentic/payagentic-go"
)

_, err := client.OpenAPI.GetWalletWithResponse(ctx, "wal_xyz")

var unauth *payagentic.UnauthorizedError
var notFound *payagentic.NotFoundError
switch {
case errors.As(err, &unauth):
    // re-auth or surface to the user
case errors.As(err, &notFound):
    // wallet doesn't exist
}

if payagentic.IsRetryable(err) {
    // retry your higher-level operation
}
```

### Re-exported types

Common generated response shapes are re-exported under the top-level `payagentic` package so you don't need to import `internal/openapi` directly:

```go
var wallet payagentic.WalletResponse
var payment payagentic.PaymentResponse
var balance payagentic.BalanceResponse
```

### Configuration options

```go
client, _ := payagentic.NewClient(
    payagentic.WithAPIKey("pa_test_…"),
    payagentic.WithAgentID("agent_…"),
    payagentic.WithBaseURL("https://api.staging.payagentic.ai"),
    payagentic.WithTimeout(30 * time.Second),
    payagentic.WithRetryPolicy(middleware.RetryPolicy{
        MaxRetries: 3,
        BaseDelay:  200 * time.Millisecond,
        MaxDelay:   5 * time.Second,
        Jitter:     0.2,
    }),
    payagentic.WithHTTPClient(myCustomClient), // overrides all middleware
)
```

## Generated transport

The HTTP transport in `internal/openapi/openapi.gen.go` is generated by [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) from the gateway's OpenAPI spec at [`../../api/openapi.json`](../../api/openapi.json). Regenerated by the monorepo's `just gen-sdk-go`.

The generator runs against the spec via a `jq` preprocessor that downgrades OpenAPI 3.1 nullable arrays (`type: [X, "null"]`) to 3.0 `nullable: true` form, since `oapi-codegen` doesn't yet support 3.1.

`internal/openapi/openapi.gen.go` is marked `linguist-generated=true`. Don't edit it directly — modify the gateway's `#[utoipa::path]` annotations + regenerate.

## Development (from the monorepo)

```bash
just gen-sdk-go    # regenerate from api/openapi.json
cd sdks/go
go vet ./...
go test ./...
go build ./...
```

The legacy `cmd/payagentic` CLI is gated behind a `legacy_cli` build tag (it predates the codegen rewrite). Build it with `go build -tags legacy_cli ./cmd/payagentic` if needed; a rewrite on the new façade is tracked as a follow-up.

## Publishing

```bash
just release-go 0.1.0   # subtree split + force-push to payagentic-go + tag v0.1.0
```

The monorepo's `sdks/go/` is the editorial source of truth. `release-go.sh` extracts that subtree into a flat tree and force-pushes to `git@github.com:payagentic/payagentic-go.git`. Consumers always `go get github.com/payagentic/payagentic-go@v0.1.0`.

The empty `payagentic-go` repo must exist on GitHub before the first release.

## License

MIT
