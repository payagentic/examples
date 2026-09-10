package middleware

import (
	"net/http"
)

// TransportOptions configures NewTransport.
type TransportOptions struct {
	// APIKey is the bearer token sent on every request.
	APIKey string
	// AgentID, when non-empty, is sent as X-Agent-ID.
	AgentID string
	// RetryPolicy controls retry behaviour. Zero value uses DefaultRetryPolicy().
	RetryPolicy RetryPolicy
	// Inner is the wrapped RoundTripper. Defaults to http.DefaultTransport.
	Inner http.RoundTripper
}

// NewTransport returns a fully wired http.RoundTripper composing:
//
//  1. authTransport — injects Authorization, X-Agent-ID, Idempotency-Key
//  2. problemDetailsTransport — converts non-2xx to typed errors
//  3. retryTransport — retries 5xx/429/network failures (sees raw status)
//  4. (inner) http.DefaultTransport — the actual network
//
// retryTransport sits innermost (below problem-details) so it inspects raw
// HTTP status codes via isRetryableStatus rather than typed errors; this
// avoids retrying 401/403/422 and only retries 408/429/502/503/504.
//
// Drop this into &http.Client{Transport: NewTransport(...)} and the
// resulting Client behaves like a vanilla http.Client but with all
// PayAgentic SDK invariants honoured.
func NewTransport(opts TransportOptions) http.RoundTripper {
	inner := opts.Inner
	if inner == nil {
		inner = http.DefaultTransport
	}
	policy := opts.RetryPolicy
	if policy.MaxRetries == 0 && policy.BaseDelay == 0 && policy.MaxDelay == 0 {
		policy = DefaultRetryPolicy()
	}

	// Innermost: retry on retryable raw HTTP status codes.
	retry := newRetryTransport(inner, policy)
	// Middle: convert any remaining non-2xx response to a typed error.
	pd := NewProblemDetailsTransport(retry)
	// Outermost: inject auth headers.
	auth := NewAuthTransport(AuthOptions{
		APIKey:  opts.APIKey,
		AgentID: opts.AgentID,
		Inner:   pd,
	})
	return auth
}

// retryTransport wraps an inner RoundTripper with WithRetry.
type retryTransport struct {
	inner  http.RoundTripper
	policy RetryPolicy
}

func newRetryTransport(inner http.RoundTripper, policy RetryPolicy) http.RoundTripper {
	return &retryTransport{inner: inner, policy: policy}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	return WithRetry(ctx, t.policy, func() (*http.Response, error) {
		// Clone the request per attempt so retried sends start from a clean
		// header set (the net/http RoundTripper contract forbids mutating
		// the original request).
		attemptReq := req.Clone(ctx)
		return t.inner.RoundTrip(attemptReq)
	})
}
