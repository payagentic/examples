package payagentic

// x402 payment-flow walker.
//
// Handles the HTTP 402 Payment Required retry loop: when a resource returns
// 402, the walker proposes a payment via the PayAgentic gateway, waits for
// settlement, then retries the original request with a payment receipt.
//
// Gateway-bound calls (`POST /v1/payments/propose`, `GET /v1/payments/{id}`)
// delegate to the generated *openapi.ClientWithResponses. The two
// mandate endpoints used by the agent kit (`/v1/mandates/issue` and
// `/v1/mandates/{jti}/revoke`) are owned by the identity service and NOT
// in the gateway spec; the MandatesService in mandate.go keeps its
// direct-HTTP implementation until those endpoints are surfaced.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/payagentic/payagentic-go/internal/middleware"
	"github.com/payagentic/payagentic-go/internal/openapi"
)

// PendingApprovalError reports a payment parked for an operator decision.
type PendingApprovalError struct{ TransactionID string }

func (e *PendingApprovalError) Error() string {
	return fmt.Sprintf("x402 payment %s requires operator approval", e.TransactionID)
}

// PolicyDeniedError reports a policy or operator denial.
type PolicyDeniedError struct {
	TransactionID string
	Cause         error
}

func (e *PolicyDeniedError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("x402 payment %s denied: %v", e.TransactionID, e.Cause)
	}
	return fmt.Sprintf("x402 payment %s denied", e.TransactionID)
}

func (e *PolicyDeniedError) Unwrap() error { return e.Cause }

// X402PaymentRequest describes a payment triggered by an HTTP 402 response.
//
// Retained as a thin public input type so callers can stage propose
// calls without importing the generated package.
type X402PaymentRequest struct {
	// URL is the API endpoint that returned HTTP 402.
	URL string `json:"url"`
	// PaymentRequirements is the value of the upstream `X-PAYMENT-REQUIREMENTS`
	// (or `X-Payment-Requirements`) header from the 402 response.
	PaymentRequirements string `json:"payment_requirements"`
	// MaxAmount is the maximum amount the caller is willing to pay (optional cap).
	MaxAmount string `json:"max_amount,omitempty"`
}

// Public re-exports of the generated x402 transport types so callers don't
// have to reach into internal/openapi.
type (
	// ProposePaymentBody is the request body for POST /v1/payments/propose.
	ProposePaymentBody = openapi.ProposePaymentJSONRequestBody
	// PaymentIntent is the gateway's PaymentResponse — returned by both
	// propose and status-poll endpoints.
	PaymentIntent = openapi.PaymentResponse
)

// Tunables for the 402 fetch loop.
const (
	x402MaxPollAttempts = 30
	x402PollInterval    = 200 * time.Millisecond
	// Approval is operator-driven, so poll less aggressively than settlement.
	x402ApprovalPollInterval = 1 * time.Second

	x402PaymentStatusOK         = "settled"
	x402PaymentStatusFail       = "failed"
	x402PaymentStatusAuthorized = "authorized"
	x402PaymentStatusDenied     = "denied"
	x402PaymentStatusRejected   = "rejected"
	x402PaymentStatusExpired    = "expired"
)

// X402Client is the x402 paywall walker. Construct via NewX402Client(*Client).
//
// Gateway calls (propose, status, settlement) go through the parent Client's
// authenticated transport. The outer fetch of the paid resource is a THIRD
// PARTY seller, so it uses a plain HTTP client instead: it must neither carry
// our gateway Bearer token nor run the gateway's RFC 7807 error mapping (which
// would turn the seller's 402 challenge into an error before we can pay it).
type X402Client struct {
	client     *Client
	sellerHTTP *http.Client
}

// NewX402Client creates a new X402Client from an existing Client.
func NewX402Client(c *Client) *X402Client {
	return &X402Client{client: c, sellerHTTP: &http.Client{}}
}

// Propose issues POST /v1/payments/propose via the generated transport
// and returns the parsed PaymentResponse.
//
// Use this directly when you already have the payment_requirements blob
// and want to drive the polling loop yourself. For the full 402 walk use
// Fetch.
func (x *X402Client) Propose(ctx context.Context, req ProposePaymentBody) (*PaymentIntent, error) {
	resp, err := x.client.OpenAPI.ProposePaymentWithResponse(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("x402 propose: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf(
			"x402 propose: unexpected status %d", resp.StatusCode())
	}
	return resp.JSON200, nil
}

// GetPayment issues GET /v1/payments/{id} via the generated transport.
//
// Used by Fetch to poll for settlement; exposed publicly so callers
// driving their own polling loop can reuse the same path.
func (x *X402Client) GetPayment(ctx context.Context, id string) (*PaymentIntent, error) {
	resp, err := x.client.OpenAPI.GetPaymentWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("x402 get payment %s: %w", id, err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf(
			"x402 get payment %s: unexpected status %d", id, resp.StatusCode())
	}
	return resp.JSON200, nil
}

// Fetch performs an x402 walk: GET (or other method) the resource; if it
// returns 402, propose a payment, poll until it settles, then retry the
// original request with the payment receipt header set.
//
// The walker uses the Client's underlying http.Client so the same retry +
// auth middleware applies to the outer fetch as to gateway calls.
func (x *X402Client) Fetch(ctx context.Context, url string, opts ...FetchOption) (*http.Response, error) {
	cfg := fetchConfig{method: http.MethodGet, maxAmount: ""}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.idempotencyKey == "" {
		cfg.idempotencyKey = middleware.GenerateIdempotencyKey()
	}

	resp, err := x.doRequest(ctx, cfg.method, url, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil
	}
	// Preserve a flat JSON body challenge before closing the response. Header
	// challenges retain precedence when both forms are present.
	paymentRequired := resp.Header.Get("Payment-Required")
	bodyChallenge, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("x402 read 402 response body: %w", readErr)
	}

	// x402 v2 (canonical spec): the challenge rides in the Payment-Required
	// header. Propose it, retry with the rail credential header the gateway
	// returns, then report the seller's Payment-Response receipt.
	if paymentRequired != "" {
		return x.fetchV2(ctx, cfg.method, url, paymentRequired, cfg.maxAmount, cfg.idempotencyKey, cfg.waitForApproval)
	}
	if resp.Header.Get("X-PAYMENT-REQUIREMENTS") == "" && resp.Header.Get("X-Payment-Requirements") == "" {
		return x.fetchBodyChallenge(ctx, cfg.method, url, bodyChallenge, cfg.maxAmount, cfg.idempotencyKey, cfg.waitForApproval)
	}

	// Legacy x402: body-derived requirements, buyer broadcasts, poll to settled.
	payment, err := x.handle402(ctx, resp, url, cfg.maxAmount, cfg.idempotencyKey)
	if err != nil {
		return nil, err
	}
	return x.doRequest(ctx, cfg.method, url, http.Header{
		"X-Payment-Receipt": []string{payment.Id.String()},
	})
}

// v2ProposeResponse is the subset of the gateway propose reply the v2 walk
// needs: the rail-formatted credential header and the transaction id. Parsed
// directly (not via the generated transport, which predates these fields).
type v2ProposeResponse struct {
	ID               string          `json:"id"`
	TransactionID    string          `json:"transaction_id"`
	PaymentSignature string          `json:"payment_signature"`
	Authorization    json.RawMessage `json:"authorization"`
	Intent           *struct {
		ID string `json:"id"`
	} `json:"intent"`
	Credential *struct {
		HeaderName  string `json:"header_name"`
		HeaderValue string `json:"header_value"`
	} `json:"credential"`
}

// fetchBodyChallenge runs the flat JSON-body x402 walk used by the staging
// vendor and older x402 sellers. The gateway signs the EIP-3009 authorization;
// the buyer retries with the canonical Authorization: Payment credential.
func (x *X402Client) fetchBodyChallenge(
	ctx context.Context, method, url string, challenge []byte, maxAmount, idempotencyKey string, waitForApproval time.Duration,
) (*http.Response, error) {
	if x.client.config.WalletID == "" {
		return nil, fmt.Errorf("x402 body challenge requires a configured wallet — construct the client with WithWalletID")
	}
	var decoded map[string]any
	if len(challenge) == 0 || json.Unmarshal(challenge, &decoded) != nil || decoded == nil {
		return nil, errors.New("x402: 402 response contained neither a challenge header nor a valid JSON object")
	}
	canonicalChallenge, err := json.Marshal(decoded)
	if err != nil {
		return nil, fmt.Errorf("x402 encode body challenge: %w", err)
	}
	body := proposeWithWallet{
		PaymentRequirements: string(canonicalChallenge),
		ResourceURL:         url,
		WalletID:            x.client.config.WalletID,
		IdempotencyKey:      idempotencyKey,
	}
	if maxAmount != "" {
		body.MaxAmount = &maxAmount
	}
	var prop v2ProposeResponse
	if err := x.client.do(ctx, http.MethodPost, "/v1/payments/propose", body, &prop); err != nil {
		if middleware.IsForbidden(err) {
			return nil, &PolicyDeniedError{TransactionID: prop.txnID(), Cause: err}
		}
		return nil, fmt.Errorf("x402 body challenge propose: %w", err)
	}
	if prop.PaymentSignature == "" {
		if waitForApproval > 0 && prop.txnID() != "" {
			return x.resumeAfterApproval(ctx, method, url, prop.txnID(), waitForApproval)
		}
		return nil, &PendingApprovalError{TransactionID: prop.txnID()}
	}

	var credentialPayload any
	if len(prop.Authorization) > 0 && string(prop.Authorization) != "null" {
		credentialPayload = struct {
			Scheme        string          `json:"scheme"`
			Authorization json.RawMessage `json:"authorization"`
			Signature     string          `json:"signature"`
		}{"exact", prop.Authorization, prop.PaymentSignature}
	} else if prop.Intent != nil && prop.Intent.ID != "" {
		credentialPayload = struct {
			Signature string `json:"signature"`
			IntentID  string `json:"intent_id"`
		}{prop.PaymentSignature, prop.Intent.ID}
	} else {
		return nil, errors.New("x402 body challenge propose returned a signature without authorization or intent")
	}
	rawCredential, err := json.Marshal(credentialPayload)
	if err != nil {
		return nil, fmt.Errorf("x402 encode payment credential: %w", err)
	}
	return x.retryWithCredential(
		ctx, method, url, "Authorization", "Payment "+base64.StdEncoding.EncodeToString(rawCredential), prop.txnID(),
	)
}

// txnID returns the transaction id to use for settlement reporting, preferring
// the explicit transaction_id and falling back to id.
func (r *v2ProposeResponse) txnID() string {
	if r.TransactionID != "" {
		return r.TransactionID
	}
	return r.ID
}

// fetchV2 runs the x402 v2 (seller-settles) walk: propose the raw challenge,
// retry the original request with the gateway-formatted credential header, then
// forward the seller's settlement receipt (advisory).
func (x *X402Client) fetchV2(
	ctx context.Context, method, url, challenge, maxAmount, idempotencyKey string, waitForApproval time.Duration,
) (*http.Response, error) {
	// The gateway requires a wallet id for v2 propose; fail with a clear message
	// rather than posting wallet_id:"" and surfacing an opaque UUID parse error.
	if x.client.config.WalletID == "" {
		return nil, fmt.Errorf("x402 v2 requires a configured wallet — construct the client with WithWalletID")
	}
	body := proposeWithWallet{
		PaymentRequirements: challenge,
		ResourceURL:         url,
		WalletID:            x.client.config.WalletID,
		IdempotencyKey:      idempotencyKey,
	}
	if maxAmount != "" {
		body.MaxAmount = &maxAmount
	}
	var prop v2ProposeResponse
	if err := x.client.do(ctx, http.MethodPost, "/v1/payments/propose", body, &prop); err != nil {
		if middleware.IsForbidden(err) {
			return nil, &PolicyDeniedError{Cause: err}
		}
		return nil, fmt.Errorf("x402 v2 propose: %w", err)
	}

	// No credential: either the payment was parked for operator approval, or the
	// server isn't on the v2 rail. If the caller opted into waiting and we have a
	// transaction to poll, resume once it's approved; otherwise fail clearly.
	if prop.Credential == nil || prop.Credential.HeaderName == "" || prop.Credential.HeaderValue == "" {
		if waitForApproval > 0 && prop.txnID() != "" {
			return x.resumeAfterApproval(ctx, method, url, prop.txnID(), waitForApproval)
		}
		return nil, &PendingApprovalError{TransactionID: prop.txnID()}
	}

	return x.retryWithCredential(ctx, method, url, prop.Credential.HeaderName, prop.Credential.HeaderValue, prop.txnID())
}

// retryWithCredential retries the resource with the rail credential header set,
// then forwards the seller's Payment-Response receipt (advisory) on success.
func (x *X402Client) retryWithCredential(
	ctx context.Context, method, url, headerName, headerValue, txnID string,
) (*http.Response, error) {
	resp, err := x.doRequest(ctx, method, url, http.Header{headerName: []string{headerValue}})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if receipt := resp.Header.Get("Payment-Response"); receipt != "" && txnID != "" {
			x.reportSettlement(ctx, txnID, receipt)
		}
	}
	return resp, nil
}

// resumeAfterApproval polls the payment until a spend-policy approval resolves:
// once it is 'authorized' with a signed credential we retry the seller with it;
// a terminal denial/expiry returns an error; the wait is bounded by `timeout`.
// Re-proposing after approval would just park a new payment behind the same
// stateless approval rule, so we fetch the parked credential and resume instead.
func (x *X402Client) resumeAfterApproval(
	ctx context.Context, method, url, txnID string, timeout time.Duration,
) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		status, headerName, headerValue, err := x.getPaymentV2(ctx, txnID)
		if err == nil {
			switch status {
			case x402PaymentStatusAuthorized:
				if headerName != "" && headerValue != "" {
					return x.retryWithCredential(ctx, method, url, headerName, headerValue, txnID)
				}
			case x402PaymentStatusDenied, x402PaymentStatusRejected,
				x402PaymentStatusExpired, x402PaymentStatusFail:
				return nil, &PolicyDeniedError{
					TransactionID: txnID,
					Cause:         errors.New("operator denied or payment expired"),
				}
			}
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("x402 v2: approval wait for payment %s ended: %w", txnID, ctx.Err())
		case <-time.After(x402ApprovalPollInterval):
		}
	}
}

// v2PaymentStatus is the subset of GET /v1/payments/{id} the resume path reads.
type v2PaymentStatus struct {
	Status     string `json:"status"`
	Credential *struct {
		HeaderName  string `json:"header_name"`
		HeaderValue string `json:"header_value"`
	} `json:"credential"`
}

// getPaymentV2 reads a payment's status and (when authorized) its rail
// credential directly, since the generated PaymentResponse predates the field.
func (x *X402Client) getPaymentV2(ctx context.Context, id string) (status, headerName, headerValue string, err error) {
	var p v2PaymentStatus
	if err := x.client.do(ctx, http.MethodGet, "/v1/payments/"+id, nil, &p); err != nil {
		return "", "", "", err
	}
	if p.Credential != nil {
		return p.Status, p.Credential.HeaderName, p.Credential.HeaderValue, nil
	}
	return p.Status, "", "", nil
}

// reportSettlement forwards the seller's base64 Payment-Response header value to
// the gateway settlement endpoint. Best-effort: errors are swallowed because the
// payments watcher independently verifies settlement on-chain.
func (x *X402Client) reportSettlement(ctx context.Context, transactionID, paymentResponse string) {
	body := map[string]string{"payment_response": paymentResponse}
	path := "/v1/payments/" + transactionID + "/settlement"
	_ = x.client.do(ctx, http.MethodPost, path, body, nil)
}

// FetchOption configures a Fetch call.
type FetchOption func(*fetchConfig)

type fetchConfig struct {
	method          string
	maxAmount       string
	idempotencyKey  string
	waitForApproval time.Duration
}

// WithMethod sets the HTTP method for the outer fetch (default GET).
func WithMethod(method string) FetchOption {
	return func(c *fetchConfig) { c.method = method }
}

// WithMaxAmount sets a cap on what the walker is willing to spend
// during the 402 walk (decimal string, e.g. "1.00").
func WithMaxAmount(amount string) FetchOption {
	return func(c *fetchConfig) { c.maxAmount = amount }
}

// WithIdempotencyKey sets a stable key for the gateway payment mutation.
// Reuse it when resuming the same parked payment.
func WithIdempotencyKey(key string) FetchOption {
	return func(c *fetchConfig) { c.idempotencyKey = key }
}

// WithWaitForApproval makes an x402 v2 walk block on a payment that a spend
// policy parked for operator approval, polling until it is approved (then
// resuming with the signed credential) or terminally denied, up to `timeout`.
// Without it, a payment needing approval returns an error instead of blocking.
func WithWaitForApproval(timeout time.Duration) FetchOption {
	return func(c *fetchConfig) { c.waitForApproval = timeout }
}

// proposeWithWallet is the propose body extended with wallet_id. The
// generated ProposePaymentRequest omits wallet_id (it's derived from the
// authenticated context in the OpenAPI spec), but the gateway requires
// it, so when a wallet is configured we post this shape directly through
// the authenticated transport instead of the generated operation.
type proposeWithWallet struct {
	PaymentRequirements string  `json:"payment_requirements"`
	ResourceURL         string  `json:"resource_url"`
	MaxAmount           *string `json:"max_amount,omitempty"`
	WalletID            string  `json:"wallet_id"`
	IdempotencyKey      string  `json:"idempotency_key,omitempty"`
}

// proposePayment issues the propose call, including wallet_id when the
// client has one configured. Falls back to the generated transport when
// no wallet is set (preserving prior behaviour).
func (x *X402Client) proposePayment(
	ctx context.Context, requirements, url, maxAmount, idempotencyKey string,
) (*PaymentIntent, error) {
	if x.client.config.WalletID != "" {
		body := proposeWithWallet{
			PaymentRequirements: requirements,
			ResourceURL:         url,
			WalletID:            x.client.config.WalletID,
			IdempotencyKey:      idempotencyKey,
		}
		if maxAmount != "" {
			body.MaxAmount = &maxAmount
		}
		var intent PaymentIntent
		if err := x.client.do(ctx, http.MethodPost, "/v1/payments/propose", body, &intent); err != nil {
			return nil, fmt.Errorf("x402 propose: %w", err)
		}
		return &intent, nil
	}

	body := ProposePaymentBody{
		PaymentRequirements: requirements,
		ResourceUrl:         url,
		IdempotencyKey:      idempotencyKey,
	}
	if maxAmount != "" {
		body.MaxAmount = &maxAmount
	}
	return x.Propose(ctx, body)
}

func (x *X402Client) doRequest(
	ctx context.Context, method, url string, extraHeaders http.Header,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("x402 build request %s %s: %w", method, url, err)
	}
	for k, vs := range extraHeaders {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := x.sellerHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("x402 fetch %s %s: %w", method, url, err)
	}
	return resp, nil
}

func (x *X402Client) handle402(
	ctx context.Context, resp *http.Response, url, maxAmount, idempotencyKey string,
) (*PaymentIntent, error) {
	requirements := resp.Header.Get("X-PAYMENT-REQUIREMENTS")
	if requirements == "" {
		requirements = resp.Header.Get("X-Payment-Requirements")
	}

	payment, err := x.proposePayment(ctx, requirements, url, maxAmount, idempotencyKey)
	if err != nil {
		return nil, err
	}
	paymentID := payment.Id.String()

	for attempt := 0; attempt < x402MaxPollAttempts; attempt++ {
		current, err := x.GetPayment(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		switch current.Status {
		case x402PaymentStatusOK:
			return current, nil
		case x402PaymentStatusFail:
			return nil, fmt.Errorf("x402 payment %s failed", paymentID)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(x402PollInterval):
		}
	}
	return nil, fmt.Errorf(
		"x402 payment %s did not settle within %d attempts",
		paymentID, x402MaxPollAttempts)
}
