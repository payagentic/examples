package payagentic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/payagentic/payagentic-go/internal/openapi"
)

const (
	maxCapabilityInvocationBytes = 64 * 1024
	maxCapabilitySuccessBytes    = maxCapabilityInvocationBytes + 16*1024
	maxCapabilityErrorBytes      = 16 * 1024
	maxCapabilityKeyBytes        = 255
	maxCapabilitySegmentBytes    = 255
	maxCapabilityTimeout         = 120 * time.Second
)

// InvokeCapabilityOptions contains caller-owned authority for one logical
// invocation. Reuse the same key only when retrying the same input.
type InvokeCapabilityOptions struct {
	IdempotencyKey string
}

// CapabilityInput is the explicit JSON object accepted by capability
// invocation. Nested values may be nil, bool, valid UTF-8 string, finite
// float64, exactly binary64-representable native integers and json.Number
// values of at most 2 KiB, []any, or map[string]any. The SDK copies containers
// before encoding. Structs, typed containers, json.RawMessage, and custom
// marshalers are rejected.
type CapabilityInput map[string]any

// CapabilityInvocationReceipt is the receipt reference in a public success envelope.
type CapabilityInvocationReceipt struct {
	ID string `json:"id"`
}

// CapabilityTestAccounting labels the non-redeemable test-mode settlement.
type CapabilityTestAccounting struct {
	Currency string `json:"currency"`
	Label    string `json:"label"`
}

// CapabilityInvocationResult is the manifest-specific public REST success envelope.
// Output remains raw JSON because its schema belongs to the immutable capability manifest.
type CapabilityInvocationResult struct {
	InvocationID      string                      `json:"invocationId"`
	Status            string                      `json:"status"`
	Output            json.RawMessage             `json:"output"`
	Receipt           CapabilityInvocationReceipt `json:"receipt"`
	Transport         string                      `json:"transport"`
	TestAccounting    CapabilityTestAccounting    `json:"testAccounting"`
	DeprecationNotice *string                     `json:"deprecationNotice,omitempty"`
}

type capabilityProblem struct {
	Code          string `json:"code"`
	Status        int    `json:"status"`
	Title         string `json:"title"`
	CorrelationID string `json:"correlationId"`
	Retry         string `json:"retry"`
	RecoveryHint  string `json:"recoveryHint"`
}

// CapabilityInvocationError is a bounded, sanitized public invocation failure.
type CapabilityInvocationError struct {
	StatusCode    int
	Code          string
	CorrelationID string
	Retry         string
	RecoveryHint  string
}

// Error implements error without including raw response bodies, URLs, or credentials.
func (e *CapabilityInvocationError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("payagentic: capability invocation failed (HTTP %d)", e.StatusCode)
	}
	return fmt.Sprintf(
		"payagentic: capability invocation failed (%s, HTTP %d)",
		e.Code,
		e.StatusCode,
	)
}

// CapabilityClient invokes manifest-specific public capability routes.
// Control-plane /v1 resources remain on Client.OpenAPI.
type CapabilityClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	timeout    time.Duration
}

func newCapabilityClient(
	baseURL string,
	apiKey string,
	httpClient *http.Client,
	timeout time.Duration,
) *CapabilityClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	boundedClient := *httpClient
	boundedClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &CapabilityClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &boundedClient,
		timeout:    timeout,
	}
}

func (c *CapabilityClient) configuration() (string, error) {
	origin, err := capabilityOrigin(c.baseURL)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(c.apiKey) == "" || containsHTTPControl(c.apiKey) {
		return "", fmt.Errorf("payagentic: capability client requires a valid in-memory API key")
	}
	if c.timeout <= 0 || c.timeout > maxCapabilityTimeout {
		return "", fmt.Errorf("payagentic: capability timeout must be between 0 and 120 seconds")
	}
	return origin, nil
}

func capabilityOrigin(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("payagentic: capability base URL must be an absolute HTTP(S) origin")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf(
			"payagentic: capability base URL must not contain credentials, a path, query, or fragment",
		)
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

func containsHTTPControl(value string) bool {
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return true
		}
	}
	return false
}

func validateCapabilitySegment(value, name string) error {
	if strings.TrimSpace(value) == "" || containsHTTPControl(value) || len(value) > maxCapabilitySegmentBytes || !utf8.ValidString(value) {
		return fmt.Errorf(
			"payagentic: %s must be a nonblank path segment of at most 255 UTF-8 bytes",
			name,
		)
	}
	return nil
}

func validateCapabilityKey(value string) error {
	if strings.TrimSpace(value) == "" || containsHTTPControl(value) || len(value) > maxCapabilityKeyBytes || !utf8.ValidString(value) {
		return fmt.Errorf(
			"payagentic: idempotency key must be nonblank and at most 255 UTF-8 bytes without control characters",
		)
	}
	return nil
}

func capabilityInput(input CapabilityInput) ([]byte, error) {
	body, err := canonicalCapabilityInputBytes(input)
	if err != nil {
		return nil, fmt.Errorf("payagentic: capability input must be JSON serializable")
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, fmt.Errorf("payagentic: capability input must serialize to a JSON object")
	}
	if len(body) > maxCapabilityInvocationBytes {
		return nil, fmt.Errorf("payagentic: capability input must not exceed 64 KiB of JSON")
	}
	return body, nil
}

// Invoke calls the pinned public capability with a caller-owned idempotency key.
// Context cancellation and the configured SDK timeout both bound the request.
func (c *CapabilityClient) Invoke(
	ctx context.Context,
	providerSlug string,
	capabilitySlug string,
	input CapabilityInput,
	options InvokeCapabilityOptions,
) (*CapabilityInvocationResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("payagentic: capability invocation context is required")
	}
	if err := validateCapabilitySegment(providerSlug, "provider slug"); err != nil {
		return nil, err
	}
	if err := validateCapabilitySegment(capabilitySlug, "capability slug"); err != nil {
		return nil, err
	}
	if err := validateCapabilityKey(options.IdempotencyKey); err != nil {
		return nil, err
	}
	origin, err := c.configuration()
	if err != nil {
		return nil, err
	}
	body, err := capabilityInput(input)
	if err != nil {
		return nil, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	path := "/c/" + url.PathEscape(providerSlug) + "/" + url.PathEscape(capabilitySlug) + "/invoke"
	request, err := http.NewRequestWithContext(
		requestCtx,
		http.MethodPost,
		origin+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("payagentic: creating bounded capability request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", options.IdempotencyKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, &CapabilityInvocationError{StatusCode: 0}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		if responseMediaType(response) != "application/json" {
			return nil, &CapabilityInvocationError{StatusCode: http.StatusOK}
		}
		raw, ok := readCapabilityBody(response, maxCapabilitySuccessBytes)
		if !ok {
			return nil, &CapabilityInvocationError{StatusCode: http.StatusOK}
		}
		return parseCapabilitySuccess(raw)
	}

	mediaType := responseMediaType(response)
	var problem *capabilityProblem
	if mediaType == "application/json" || mediaType == "application/problem+json" {
		if raw, ok := readCapabilityBody(response, maxCapabilityErrorBytes); ok {
			problem = parseCapabilityProblem(raw, response.StatusCode, c.apiKey)
		}
	}
	return nil, capabilityError(response.StatusCode, problem)
}

func responseMediaType(response *http.Response) string {
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return ""
	}
	return strings.ToLower(mediaType)
}

func readCapabilityBody(response *http.Response, maximum int64) ([]byte, bool) {
	if response.ContentLength < -1 || response.ContentLength > maximum {
		return nil, false
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maximum+1))
	if err != nil || int64(len(body)) > maximum {
		return nil, false
	}
	return body, true
}

func decodeCapabilityJSON(raw []byte, target any) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	return decoder.Decode(&struct{}{}) == io.EOF
}

func parseCapabilitySuccess(raw []byte) (*CapabilityInvocationResult, error) {
	var result CapabilityInvocationResult
	if !decodeCapabilityJSON(raw, &result) ||
		!validUUID(result.InvocationID) ||
		result.Status != "settled" ||
		!jsonObject(result.Output) ||
		len(result.Output) > maxCapabilityInvocationBytes ||
		!validUUID(result.Receipt.ID) ||
		result.Transport != "rest" && result.Transport != "mcp" ||
		result.TestAccounting.Currency != "USDC" ||
		result.TestAccounting.Label != openapi.CapabilityTestAccountingLabel {
		return nil, &CapabilityInvocationError{StatusCode: http.StatusOK}
	}
	return &result, nil
}

func jsonObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == '{' && json.Valid(trimmed)
}

func parseCapabilityProblem(raw []byte, status int, apiKey string) *capabilityProblem {
	var problem capabilityProblem
	if !decodeCapabilityJSON(raw, &problem) ||
		problem.Status != status ||
		problem.Title != "Capability invocation failed" ||
		strings.Contains(problem.Code, apiKey) ||
		!validUUID(problem.CorrelationID) ||
		len(problem.RecoveryHint) > 256 ||
		strings.Contains(problem.RecoveryHint, apiKey) {
		return nil
	}
	for _, tuple := range openapi.CapabilityProblemTuples {
		if tuple.Code == problem.Code &&
			tuple.Status == problem.Status &&
			tuple.Retry == problem.Retry &&
			tuple.RecoveryHint == problem.RecoveryHint {
			return &problem
		}
	}
	return nil
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.String() == strings.ToLower(value) && parsed.Variant() == uuid.RFC4122
}

func capabilityError(status int, problem *capabilityProblem) *CapabilityInvocationError {
	invocationError := &CapabilityInvocationError{StatusCode: status}
	if problem != nil {
		invocationError.Code = problem.Code
		invocationError.CorrelationID = problem.CorrelationID
		invocationError.Retry = problem.Retry
		invocationError.RecoveryHint = problem.RecoveryHint
	}
	return invocationError
}

// NewCapabilityIdempotencyKey returns a standards-correct RFC 9562 UUIDv7.
func NewCapabilityIdempotencyKey() (string, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("payagentic: generating capability UUIDv7: %w", err)
	}
	return value.String(), nil
}
