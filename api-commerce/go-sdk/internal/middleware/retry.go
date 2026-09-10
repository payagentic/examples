package middleware

import (
	"context"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

// RetryPolicy configures the retry behaviour for transient failures.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts (0 means no retries).
	MaxRetries int
	// BaseDelay is the initial backoff delay.
	BaseDelay time.Duration
	// MaxDelay is the upper bound on backoff delay.
	MaxDelay time.Duration
	// Jitter enables randomised delay to avoid thundering herds.
	Jitter bool
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Jitter:     true,
	}
}

// isRetryableStatus reports whether the HTTP status code is retryable.
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusRequestTimeout:
		return true
	default:
		return false
	}
}

// backoff calculates the delay for a given attempt using exponential backoff.
func (p RetryPolicy) backoff(attempt int) time.Duration {
	delay := time.Duration(float64(p.BaseDelay) * math.Pow(2, float64(attempt)))
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	if p.Jitter {
		// Add up to 50% jitter.
		jitter := time.Duration(rand.Int64N(int64(delay) / 2))
		delay += jitter
	}
	return delay
}

// WithRetry executes fn with retries according to the given policy.
// Only responses with retryable HTTP status codes trigger a retry.
func WithRetry(ctx context.Context, policy RetryPolicy, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := policy.backoff(attempt - 1)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		resp, err := fn()
		if err != nil {
			lastErr = err
			// Network-level errors are retryable.
			continue
		}

		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Close the body of the retryable response so we don't leak connections.
		resp.Body.Close()
		lastResp = resp
		lastErr = nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return lastResp, nil
}
