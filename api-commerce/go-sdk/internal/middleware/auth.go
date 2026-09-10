package middleware

import (
	"net/http"
)

// AuthOptions configures NewAuthTransport.
type AuthOptions struct {
	// APIKey is sent as `Authorization: Bearer <APIKey>` on every request.
	APIKey string
	// AgentID, when non-empty, is sent as `X-Agent-ID: <AgentID>`.
	AgentID string
	// Inner is the wrapped RoundTripper. Defaults to http.DefaultTransport.
	Inner http.RoundTripper
}

// nonIdempotentMethods are HTTP methods that need an Idempotency-Key.
var nonIdempotentMethods = map[string]struct{}{
	"POST":   {},
	"PUT":    {},
	"PATCH":  {},
	"DELETE": {},
}

// authTransport is an http.RoundTripper that injects Authorization,
// optional X-Agent-ID, and Idempotency-Key on non-idempotent methods.
type authTransport struct {
	opts AuthOptions
}

// NewAuthTransport returns a RoundTripper that authenticates outgoing
// requests with the configured API key and adds an Idempotency-Key
// (UUID v4) on POST/PUT/PATCH/DELETE unless the caller supplied one.
func NewAuthTransport(opts AuthOptions) http.RoundTripper {
	if opts.Inner == nil {
		opts.Inner = http.DefaultTransport
	}
	return &authTransport{opts: opts}
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone — net/http RoundTripper contract forbids mutating the request.
	req = req.Clone(req.Context())

	req.Header.Set("Authorization", "Bearer "+t.opts.APIKey)
	if t.opts.AgentID != "" {
		req.Header.Set("X-Agent-ID", t.opts.AgentID)
	}
	if _, nonIdempotent := nonIdempotentMethods[req.Method]; nonIdempotent {
		if req.Header.Get("Idempotency-Key") == "" {
			req.Header.Set("Idempotency-Key", GenerateIdempotencyKey())
		}
	}

	return t.opts.Inner.RoundTrip(req)
}
