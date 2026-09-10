package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// problemDetailsTransport is an http.RoundTripper that converts non-2xx
// responses into typed errors via ParseErrorResponse.
type problemDetailsTransport struct {
	inner http.RoundTripper
}

// NewProblemDetailsTransport returns a RoundTripper that, after the inner
// RT returns, reads non-2xx response bodies and returns the matching typed
// error (UnauthorizedError, RateLimitError, etc) via ParseErrorResponse.
//
// 2xx responses pass through with the body intact.
func NewProblemDetailsTransport(inner http.RoundTripper) http.RoundTripper {
	if inner == nil {
		inner = http.DefaultTransport
	}
	return &problemDetailsTransport{inner: inner}
}

func (t *problemDetailsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode < 400 {
		return resp, nil
	}

	// Drain the body so ParseErrorResponse can inspect it. The response is
	// returned with a fresh body containing the same bytes so callers can
	// still read it.
	body, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	if readErr != nil {
		return resp, readErr
	}

	return resp, ParseErrorResponse(resp.StatusCode, body, resp.Header.Get("X-Trace-ID"))
}
