package payagentic

import (
	"net/http"
	"time"

	"github.com/payagentic/payagentic-go/internal/middleware"
	"github.com/payagentic/payagentic-go/internal/openapi"
)

// Config holds the configuration for the PayAgentic client.
type Config struct {
	// APIKey is the bearer token used for authentication.
	APIKey string
	// BaseURL is the PayAgentic API base URL.
	BaseURL string
	// AgentID is the optional agent identifier sent as a request header.
	AgentID string
	// WalletID is the wallet the x402 walker debits when proposing a payment.
	// The gateway's propose endpoint requires it, so a payment cannot settle
	// without a wallet configured here (via WithWalletID or the
	// PAYAGENTIC_WALLET_ID environment variable).
	WalletID string
	// Timeout is the per-request timeout applied to the underlying http.Client.
	Timeout time.Duration
}

// RetryPolicy re-exports the retry policy configuration from the
// internal middleware package so callers can stay on the parent import.
type RetryPolicy = middleware.RetryPolicy

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return middleware.DefaultRetryPolicy()
}

// ProblemDetails re-exports the RFC 9457 error envelope from the
// internal middleware package.
type ProblemDetails = middleware.ProblemDetails

// APIError re-exports the base API error type.
type APIError = middleware.APIError

// NotFoundError re-exports the 404 typed error.
type NotFoundError = middleware.NotFoundError

// UnauthorizedError re-exports the 401 typed error.
type UnauthorizedError = middleware.UnauthorizedError

// ForbiddenError re-exports the 403 typed error.
type ForbiddenError = middleware.ForbiddenError

// RateLimitError re-exports the 429 typed error.
type RateLimitError = middleware.RateLimitError

// ConflictError re-exports the 409 typed error.
type ConflictError = middleware.ConflictError

// ValidationError re-exports the 422 typed error.
type ValidationError = middleware.ValidationError

// IsRetryable reports whether the given error represents a transient failure
// that should be retried.
func IsRetryable(err error) bool { return middleware.IsRetryable(err) }

// IsNotFound reports whether the error is a 404 Not Found.
func IsNotFound(err error) bool { return middleware.IsNotFound(err) }

// IsUnauthorized reports whether the error is a 401 Unauthorized.
func IsUnauthorized(err error) bool { return middleware.IsUnauthorized(err) }

// IsForbidden reports whether the error is a 403 Forbidden.
func IsForbidden(err error) bool { return middleware.IsForbidden(err) }

// IsRateLimit reports whether the error is a 429 Too Many Requests.
func IsRateLimit(err error) bool { return middleware.IsRateLimit(err) }

// IsConflict reports whether the error is a 409 Conflict.
func IsConflict(err error) bool { return middleware.IsConflict(err) }

// IsValidation reports whether the error is a 422 Unprocessable Entity.
func IsValidation(err error) bool { return middleware.IsValidation(err) }

// Client is the PayAgentic API client. Use NewClient to create one.
//
// It exposes:
//   - OpenAPI: a generated *openapi.Client with one method per gateway endpoint
//     (e.g. client.OpenAPI.ListWalletsWithResponse(ctx, nil))
//   - Capabilities: the bounded public capability invocation helper
//   - X402: the x402 payment helper
//
// The Client's underlying http.Client has bearer auth, retries, idempotency
// keys, and RFC 7807 typed errors composed via internal/middleware.NewTransport.
type Client struct {
	config      Config
	httpClient  *http.Client
	retryPolicy RetryPolicy

	// OpenAPI exposes the generated transport with all gateway operations
	// in the *WithResponse form (status-code-typed responses).
	OpenAPI *openapi.ClientWithResponses
	// Capabilities invokes pinned public capability routes with bounded payloads.
	Capabilities *CapabilityClient
	// X402 is the x402 payment helper.
	X402 *X402Client
}

// PaginatedResponse wraps a paginated list of items returned by the API.
type PaginatedResponse[T any] struct {
	// Items contains the page of results.
	Items []T `json:"items"`
	// NextCursor is the opaque cursor for fetching the next page. Empty when there are no more pages.
	NextCursor string `json:"next_cursor,omitempty"`
	// TotalCount is the total number of items matching the query, if available.
	TotalCount int64 `json:"total_count,omitempty"`
}

// Organization represents an organization on the platform.
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents a user within an organization.
type User struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Agent represents an AI agent registered on the platform.
type Agent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	KeyFingerprint string    `json:"key_fingerprint,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Wallet represents a USDC smart account wallet.
type Wallet struct {
	ID             string       `json:"id"`
	OrganizationID string       `json:"organization_id"`
	AgentID        string       `json:"agent_id,omitempty"`
	Address        string       `json:"address"`
	Chain          string       `json:"chain"`
	Balance        string       `json:"balance"`
	Status         WalletStatus `json:"status"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// Transaction represents a blockchain transaction.
type Transaction struct {
	ID        string            `json:"id"`
	WalletID  string            `json:"wallet_id"`
	Type      TransactionType   `json:"type"`
	Status    TransactionStatus `json:"status"`
	Amount    string            `json:"amount"`
	Currency  string            `json:"currency"`
	TxHash    string            `json:"tx_hash,omitempty"`
	Chain     string            `json:"chain"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Payment represents an x402 payment.
type Payment struct {
	ID            string        `json:"id"`
	WalletID      string        `json:"wallet_id"`
	AgentID       string        `json:"agent_id"`
	Status        PaymentStatus `json:"status"`
	Amount        string        `json:"amount"`
	Currency      string        `json:"currency"`
	Recipient     string        `json:"recipient"`
	Description   string        `json:"description,omitempty"`
	TransactionID string        `json:"transaction_id,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// Policy represents a spend policy governing agent wallets.
type Policy struct {
	ID             string      `json:"id"`
	OrganizationID string      `json:"organization_id"`
	Name           string      `json:"name"`
	Type           PolicyType  `json:"type"`
	Rules          PolicyRules `json:"rules"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// PolicyRules defines the set of rules within a policy.
type PolicyRules struct {
	MaxAmount         string       `json:"max_amount,omitempty"`
	DailyLimit        string       `json:"daily_limit,omitempty"`
	MonthlyLimit      string       `json:"monthly_limit,omitempty"`
	AllowedChains     []string     `json:"allowed_chains,omitempty"`
	AllowedRecipients []string     `json:"allowed_recipients,omitempty"`
	TimeWindows       []TimeWindow `json:"time_windows,omitempty"`
}

// TimeWindow defines an allowed time window for transactions.
type TimeWindow struct {
	DaysOfWeek []string `json:"days_of_week"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
	Timezone   string   `json:"timezone"`
}

// Approval represents a pending approval request.
type Approval struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Status      ApprovalStatus `json:"status"`
	RequestedBy string         `json:"requested_by"`
	ReviewedBy  string         `json:"reviewed_by,omitempty"`
	Reason      string         `json:"reason,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AuditEntry represents an immutable audit log entry.
type AuditEntry struct {
	ID        string         `json:"id"`
	EventType string         `json:"event_type"`
	ActorID   string         `json:"actor_id"`
	ActorType string         `json:"actor_type"`
	Payload   map[string]any `json:"payload,omitempty"`
	Hash      string         `json:"hash"`
	PrevHash  string         `json:"prev_hash"`
	CreatedAt time.Time      `json:"created_at"`
}

// --- Status enums ---

// TransactionStatus represents the lifecycle status of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// TransactionType represents the type of transaction.
type TransactionType string

const (
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypeFund     TransactionType = "fund"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypePayment  TransactionType = "payment"
)

// PaymentStatus represents the lifecycle status of a payment.
type PaymentStatus string

const (
	PaymentStatusProposed  PaymentStatus = "proposed"
	PaymentStatusApproved  PaymentStatus = "approved"
	PaymentStatusExecuting PaymentStatus = "executing"
	PaymentStatusSettled   PaymentStatus = "settled"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRejected  PaymentStatus = "rejected"
)

// WalletStatus represents the status of a wallet.
type WalletStatus string

const (
	WalletStatusActive WalletStatus = "active"
	WalletStatusFrozen WalletStatus = "frozen"
	WalletStatusClosed WalletStatus = "closed"
)

// PolicyType represents the type of policy.
type PolicyType string

const (
	PolicyTypeSpendCap   PolicyType = "spend_cap"
	PolicyTypeAllowlist  PolicyType = "allowlist"
	PolicyTypeTimeWindow PolicyType = "time_window"
	PolicyTypeVelocity   PolicyType = "velocity"
	PolicyTypeComposite  PolicyType = "composite"
)

// ApprovalStatus represents the status of an approval request.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusDenied   ApprovalStatus = "denied"
	ApprovalStatusExpired  ApprovalStatus = "expired"
)

// --- Request types ---

// CreateOrganizationRequest contains the fields for creating an organization.
type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

// UpdateOrganizationRequest contains the fields for updating an organization.
type UpdateOrganizationRequest struct {
	Name string `json:"name,omitempty"`
}

// CreateAgentRequest contains the fields for creating an agent.
type CreateAgentRequest struct {
	Name string `json:"name"`
}

// FundWalletRequest contains the fields for funding a wallet.
type FundWalletRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Source   string `json:"source"`
}

// WithdrawWalletRequest contains the fields for withdrawing from a wallet.
type WithdrawWalletRequest struct {
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Destination string `json:"destination"`
}

// ProposePaymentRequest contains the fields for proposing an x402 payment.
type ProposePaymentRequest struct {
	WalletID    string `json:"wallet_id"`
	Recipient   string `json:"recipient"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description,omitempty"`
}

// CreatePolicyRequest contains the fields for creating a spend policy.
type CreatePolicyRequest struct {
	Name  string      `json:"name"`
	Type  PolicyType  `json:"type"`
	Rules PolicyRules `json:"rules"`
}

// UpdatePolicyRequest contains the fields for updating a spend policy.
type UpdatePolicyRequest struct {
	Name  string       `json:"name,omitempty"`
	Rules *PolicyRules `json:"rules,omitempty"`
}

// ApprovalDecisionRequest contains the fields for approving or denying a request.
type ApprovalDecisionRequest struct {
	Reason string `json:"reason,omitempty"`
}
