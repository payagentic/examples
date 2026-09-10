package payagentic

// Re-exports of common generated response types from the OpenAPI transport.
//
// The hand-written domain types in types.go (Wallet, Agent, Payment, Policy,
// Transaction, Approval, Organization) are kept as the ergonomic public API.
// These aliases surface the raw `*Response` shapes returned by the generated
// transport (Client.OpenAPI.*WithResponse) so callers can type-check against
// them without reaching into `internal/openapi`.
//
// Use them when working directly with the generated transport, e.g.:
//
//	resp, err := client.OpenAPI.ListWalletsWithResponse(ctx, nil)
//	if err != nil { return err }
//	var w payagentic.WalletResponse = (*resp.JSON200.Items)[0]

import "github.com/payagentic/payagentic-go/internal/openapi"

type (
	// ClientWithResponses is the generated status-code-aware API client.
	ClientWithResponses = openapi.ClientWithResponses
	// ClientWithResponsesInterface is the generated client method surface.
	ClientWithResponsesInterface = openapi.ClientWithResponsesInterface
	// RequestEditorFn customizes one generated request before transport.
	RequestEditorFn = openapi.RequestEditorFn

	// WalletResponse is the gateway response shape for a single wallet.
	WalletResponse = openapi.WalletResponse
	// AgentResponse is the gateway response shape for a single agent.
	AgentResponse = openapi.AgentResponse
	// PaymentResponse is the gateway response shape for a payment intent.
	// Also aliased as PaymentIntent in x402.go for x402-flow ergonomics.
	PaymentResponse = openapi.PaymentResponse
	// TransactionResponse is the gateway response shape for a single transaction.
	TransactionResponse = openapi.TransactionResponse
	// PolicyResponse is the gateway response shape for a single policy.
	PolicyResponse = openapi.PolicyResponse
	// OrganizationResponse is the gateway response shape for an organization.
	OrganizationResponse = openapi.OrganizationResponse
	// PayoutResponse is the gateway response shape for a merchant payout.
	PayoutResponse = openapi.PayoutResponse
	// RefundResponse is the gateway response shape for a refund.
	RefundResponse = openapi.RefundResponse
	// ApiKeyResponse is the gateway response shape for an API key.
	ApiKeyResponse = openapi.ApiKeyResponse
	// CreateApiKeyResponse is the one-time-secret payload returned on key creation.
	CreateApiKeyResponse = openapi.CreateApiKeyResponse
	// BalanceResponse is the gateway response shape for a wallet balance query.
	BalanceResponse = openapi.BalanceResponse
	// MeResponse is the gateway response shape for the /v1/me identity probe.
	MeResponse = openapi.MeResponse
	// HealthResponse is the gateway response shape for /healthz.
	HealthResponse = openapi.HealthResponse
	// StatsResponse is the gateway response shape for the merchant /stats surface.
	StatsResponse = openapi.StatsResponse
	// DashboardStats is the dashboard summary payload.
	DashboardStats = openapi.DashboardStats

	// Capability control-plane request and response models.
	ApproveCapabilityBindingRequest              = openapi.ApproveCapabilityBindingRequest
	CapabilityBindingAuthoritySummary            = openapi.CapabilityBindingAuthoritySummary
	CapabilityBindingListResponse                = openapi.CapabilityBindingListResponse
	CapabilityBindingResponse                    = openapi.CapabilityBindingResponse
	CapabilityDataPolicyInput                    = openapi.CapabilityDataPolicyInput
	CapabilityDirectoryItem                      = openapi.CapabilityDirectoryItem
	CapabilityDirectoryListResponse              = openapi.CapabilityDirectoryListResponse
	CapabilityDirectoryProvider                  = openapi.CapabilityDirectoryProvider
	CapabilityDraftListResponse                  = openapi.CapabilityDraftListResponse
	CapabilityDraftResponse                      = openapi.CapabilityDraftResponse
	CapabilityInvocationListResponse             = openapi.CapabilityInvocationListResponse
	CapabilityInvocationMetadata                 = openapi.CapabilityInvocationMetadata
	CapabilityLifecycleRequest                   = openapi.CapabilityLifecycleRequest
	CapabilityLimitsInput                        = openapi.CapabilityLimitsInput
	CapabilityPriceInput                         = openapi.CapabilityPriceInput
	CapabilityPriceResponse                      = openapi.CapabilityPriceResponse
	CapabilityReceiptResponse                    = openapi.CapabilityReceiptResponse
	CapabilityTestResponse                       = openapi.CapabilityTestResponse
	CapabilityTransportInput                     = openapi.CapabilityTransportInput
	CapabilityVersionListResponse                = openapi.CapabilityVersionListResponse
	CapabilityVersionResponse                    = openapi.CapabilityVersionResponse
	CreateCapabilityBindingRequest               = openapi.CreateCapabilityBindingRequest
	CreateCapabilityDraftRequest                 = openapi.CreateCapabilityDraftRequest
	EmptyCapabilityBindingAction                 = openapi.EmptyCapabilityBindingAction
	PublishCapabilityRequest                     = openapi.PublishCapabilityRequest
	TestCapabilityRequest                        = openapi.TestCapabilityRequest
	UpdateCapabilityDraftRequest                 = openapi.UpdateCapabilityDraftRequest
	CreateAgentCapabilityBindingParams           = openapi.CreateAgentCapabilityBindingParams
	ApproveAgentCapabilityBindingParams          = openapi.ApproveAgentCapabilityBindingParams
	PauseAgentCapabilityBindingParams            = openapi.PauseAgentCapabilityBindingParams
	RevokeAgentCapabilityBindingParams           = openapi.RevokeAgentCapabilityBindingParams
	CreateCapabilityParams                       = openapi.CreateCapabilityParams
	UpdateCapabilityParams                       = openapi.UpdateCapabilityParams
	ArchiveCapabilityParams                      = openapi.ArchiveCapabilityParams
	DeprecateCapabilityParams                    = openapi.DeprecateCapabilityParams
	PublishCapabilityParams                      = openapi.PublishCapabilityParams
	SuspendCapabilityParams                      = openapi.SuspendCapabilityParams
	TestCapabilityParams                         = openapi.TestCapabilityParams
	CreateAgentCapabilityBindingJSONRequestBody  = openapi.CreateAgentCapabilityBindingJSONRequestBody
	ApproveAgentCapabilityBindingJSONRequestBody = openapi.ApproveAgentCapabilityBindingJSONRequestBody
	PauseAgentCapabilityBindingJSONRequestBody   = openapi.PauseAgentCapabilityBindingJSONRequestBody
	RevokeAgentCapabilityBindingJSONRequestBody  = openapi.RevokeAgentCapabilityBindingJSONRequestBody
	CreateCapabilityJSONRequestBody              = openapi.CreateCapabilityJSONRequestBody
	UpdateCapabilityJSONRequestBody              = openapi.UpdateCapabilityJSONRequestBody
	ArchiveCapabilityJSONRequestBody             = openapi.ArchiveCapabilityJSONRequestBody
	DeprecateCapabilityJSONRequestBody           = openapi.DeprecateCapabilityJSONRequestBody
	PublishCapabilityJSONRequestBody             = openapi.PublishCapabilityJSONRequestBody
	SuspendCapabilityJSONRequestBody             = openapi.SuspendCapabilityJSONRequestBody
	TestCapabilityJSONRequestBody                = openapi.TestCapabilityJSONRequestBody

	// Capability control-plane operation response wrappers.
	ListAgentCapabilityBindingsHTTPResponse   = openapi.ListAgentCapabilityBindingsHTTPResponse
	ListCapabilitiesHTTPResponse              = openapi.ListCapabilitiesHTTPResponse
	CreateCapabilityHTTPResponse              = openapi.CreateCapabilityHTTPResponse
	GetCapabilityHTTPResponse                 = openapi.GetCapabilityHTTPResponse
	UpdateCapabilityHTTPResponse              = openapi.UpdateCapabilityHTTPResponse
	ArchiveCapabilityHTTPResponse             = openapi.ArchiveCapabilityHTTPResponse
	DeprecateCapabilityHTTPResponse           = openapi.DeprecateCapabilityHTTPResponse
	PublishCapabilityHTTPResponse             = openapi.PublishCapabilityHTTPResponse
	SuspendCapabilityHTTPResponse             = openapi.SuspendCapabilityHTTPResponse
	TestCapabilityHTTPResponse                = openapi.TestCapabilityHTTPResponse
	ListCapabilityVersionsHTTPResponse        = openapi.ListCapabilityVersionsHTTPResponse
	SearchCapabilitiesHTTPResponse            = openapi.SearchCapabilitiesHTTPResponse
	GetDirectoryCapabilityHTTPResponse        = openapi.GetDirectoryCapabilityHTTPResponse
	CreateAgentCapabilityBindingHTTPResponse  = openapi.CreateAgentCapabilityBindingHTTPResponse
	ApproveAgentCapabilityBindingHTTPResponse = openapi.ApproveAgentCapabilityBindingHTTPResponse
	PauseAgentCapabilityBindingHTTPResponse   = openapi.PauseAgentCapabilityBindingHTTPResponse
	RevokeAgentCapabilityBindingHTTPResponse  = openapi.RevokeAgentCapabilityBindingHTTPResponse
	ListCapabilityInvocationsHTTPResponse     = openapi.ListCapabilityInvocationsHTTPResponse
	GetCapabilityInvocationHTTPResponse       = openapi.GetCapabilityInvocationHTTPResponse
	GetCapabilityReceiptHTTPResponse          = openapi.GetCapabilityReceiptHTTPResponse
)
