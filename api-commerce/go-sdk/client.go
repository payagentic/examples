// Package payagentic is the official Go SDK for PayAgentic.
//
// Example usage:
//
//	import "github.com/payagentic/payagentic-go"
//
//	client, err := payagentic.NewClient(payagentic.WithAPIKey("pa_test_…"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	ctx := context.Background()
//	wallets, err := client.OpenAPI.ListWalletsWithResponse(ctx, nil)
//	resp, err := client.X402.Fetch(ctx, "https://vendor.example.com/data")
package payagentic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/payagentic/payagentic-go/internal/middleware"
	"github.com/payagentic/payagentic-go/internal/openapi"
)

// DefaultBaseURL is the production gateway URL used when WithBaseURL is not set.
const DefaultBaseURL = "https://app.payagentic.ai"

// DefaultTimeout is the http.Client request timeout used when WithTimeout
// is not set.
const DefaultTimeout = 30 * time.Second

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the API key sent as Bearer token.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.config.APIKey = key }
}

// WithAgentID sets the X-Agent-ID header.
func WithAgentID(id string) Option {
	return func(c *Client) { c.config.AgentID = id }
}

// WithBaseURL overrides the gateway URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.config.BaseURL = url }
}

// WithWalletID sets the wallet the x402 walker debits when proposing a
// payment. Required for x402.Fetch to settle — the gateway's propose
// endpoint rejects a proposal with no wallet.
func WithWalletID(id string) Option {
	return func(c *Client) { c.config.WalletID = id }
}

// WithHTTPClient overrides the underlying http.Client. When set, the SDK
// uses this client as-is — middleware composition is the caller's responsibility.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithRetryPolicy overrides the retry policy.
func WithRetryPolicy(p middleware.RetryPolicy) Option {
	return func(c *Client) { c.retryPolicy = p }
}

// WithTimeout sets the request timeout on the underlying http.Client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.config.Timeout = d }
}

// FromEnvironment is a no-op Option retained for backward compatibility.
// Environment resolution now happens automatically inside NewClient.
func FromEnvironment() Option {
	return func(c *Client) {}
}

// NewClient builds a Client with the given options.
//
// Resolves APIKey from PAYAGENTIC_API_KEY env, AgentID from PAYAGENTIC_AGENT_ID,
// and BaseURL from PAYAGENTIC_BASE_URL when not set explicitly via Options.
// Returns an error if no API key is resolvable.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		config: Config{
			BaseURL: DefaultBaseURL,
			Timeout: DefaultTimeout,
		},
		retryPolicy: middleware.DefaultRetryPolicy(),
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.config.APIKey == "" {
		c.config.APIKey = os.Getenv("PAYAGENTIC_API_KEY")
	}
	if c.config.AgentID == "" {
		c.config.AgentID = os.Getenv("PAYAGENTIC_AGENT_ID")
	}
	if c.config.WalletID == "" {
		c.config.WalletID = os.Getenv("PAYAGENTIC_WALLET_ID")
	}
	if envURL := os.Getenv("PAYAGENTIC_BASE_URL"); envURL != "" && c.config.BaseURL == DefaultBaseURL {
		c.config.BaseURL = envURL
	}

	if c.config.APIKey == "" {
		return nil, fmt.Errorf(
			"payagentic: API key required; pass WithAPIKey or set PAYAGENTIC_API_KEY")
	}

	capabilityHTTPClient := c.httpClient
	if capabilityHTTPClient == nil {
		capabilityHTTPClient = &http.Client{Timeout: c.config.Timeout}
	}

	if c.httpClient == nil {
		transport := middleware.NewTransport(middleware.TransportOptions{
			APIKey:      c.config.APIKey,
			AgentID:     c.config.AgentID,
			RetryPolicy: c.retryPolicy,
		})
		c.httpClient = &http.Client{
			Transport: transport,
			Timeout:   c.config.Timeout,
		}
	}

	gen, err := openapi.NewClientWithResponses(c.config.BaseURL, openapi.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, fmt.Errorf("payagentic: constructing openapi client: %w", err)
	}
	c.OpenAPI = gen
	capabilities := newCapabilityClient(
		c.config.BaseURL,
		c.config.APIKey,
		capabilityHTTPClient,
		c.config.Timeout,
	)
	c.Capabilities = capabilities
	c.X402 = NewX402Client(c)

	return c, nil
}

// Close releases any resources held by the Client.
func (c *Client) Close() error {
	return nil
}

// do executes an HTTP request with retries and JSON marshalling.
//
// Retained for legacy callers (mandate.go, x402.go) until Task 9 reworks
// them on the generated transport. The middleware composition already
// handles auth, retries, and RFC 7807 error mapping, so this method only
// needs to marshal/unmarshal JSON and surface non-2xx as typed errors.
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	url := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "payagentic-go/"+Version)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return middleware.ParseErrorResponse(resp.StatusCode, respBody, resp.Header.Get("X-Trace-ID"))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshalling response: %w", err)
		}
	}
	return nil
}
