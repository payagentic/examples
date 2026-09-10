// Package middleware composes auth, retry, and RFC 7807 error mapping
// for the PayAgentic HTTP transport.
package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ProblemDetails follows RFC 9457 for structured error responses.
type ProblemDetails struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// APIError represents an API error response from PayAgentic.
type APIError struct {
	// StatusCode is the HTTP status code returned by the API.
	StatusCode int
	// Problem contains the structured error details (RFC 9457).
	Problem ProblemDetails
	// TraceID is the trace identifier for correlating with server logs.
	TraceID string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Problem.Detail != "" {
		return fmt.Sprintf("payagentic: %d %s: %s (trace: %s)", e.StatusCode, e.Problem.Title, e.Problem.Detail, e.TraceID)
	}
	if e.Problem.Title != "" {
		return fmt.Sprintf("payagentic: %d %s (trace: %s)", e.StatusCode, e.Problem.Title, e.TraceID)
	}
	return fmt.Sprintf("payagentic: %d (trace: %s)", e.StatusCode, e.TraceID)
}

// NotFoundError is returned when a resource is not found (HTTP 404).
type NotFoundError struct{ *APIError }

// UnauthorizedError is returned when authentication fails (HTTP 401).
type UnauthorizedError struct{ *APIError }

// ForbiddenError is returned when authorization fails (HTTP 403).
type ForbiddenError struct{ *APIError }

// RateLimitError is returned when the rate limit is exceeded (HTTP 429).
type RateLimitError struct{ *APIError }

// ConflictError is returned on resource conflicts (HTTP 409).
type ConflictError struct{ *APIError }

// ValidationError is returned on invalid input (HTTP 422).
type ValidationError struct{ *APIError }

// ParseErrorResponse constructs the appropriate typed error from an HTTP response.
func ParseErrorResponse(statusCode int, body []byte, traceID string) error {
	apiErr := &APIError{
		StatusCode: statusCode,
		TraceID:    traceID,
	}

	if len(body) > 0 {
		_ = json.Unmarshal(body, &apiErr.Problem)
	}
	if apiErr.Problem.Status == 0 {
		apiErr.Problem.Status = statusCode
	}

	switch statusCode {
	case http.StatusNotFound:
		return &NotFoundError{APIError: apiErr}
	case http.StatusUnauthorized:
		return &UnauthorizedError{APIError: apiErr}
	case http.StatusForbidden:
		return &ForbiddenError{APIError: apiErr}
	case http.StatusTooManyRequests:
		return &RateLimitError{APIError: apiErr}
	case http.StatusConflict:
		return &ConflictError{APIError: apiErr}
	case http.StatusUnprocessableEntity:
		return &ValidationError{APIError: apiErr}
	default:
		return apiErr
	}
}

// IsRetryable reports whether the given error represents a transient failure
// that should be retried.
func IsRetryable(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.StatusCode {
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

// IsNotFound reports whether the error is a 404 Not Found.
func IsNotFound(err error) bool {
	var target *NotFoundError
	return errors.As(err, &target)
}

// IsUnauthorized reports whether the error is a 401 Unauthorized.
func IsUnauthorized(err error) bool {
	var target *UnauthorizedError
	return errors.As(err, &target)
}

// IsForbidden reports whether the error is a 403 Forbidden.
func IsForbidden(err error) bool {
	var target *ForbiddenError
	return errors.As(err, &target)
}

// IsRateLimit reports whether the error is a 429 Too Many Requests.
func IsRateLimit(err error) bool {
	var target *RateLimitError
	return errors.As(err, &target)
}

// IsConflict reports whether the error is a 409 Conflict.
func IsConflict(err error) bool {
	var target *ConflictError
	return errors.As(err, &target)
}

// IsValidation reports whether the error is a 422 Unprocessable Entity.
func IsValidation(err error) bool {
	var target *ValidationError
	return errors.As(err, &target)
}
