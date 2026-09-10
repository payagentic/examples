// Code generated from api/openapi.json by generate-capability-sdk-contracts.mjs. DO NOT EDIT.

package openapi

const CapabilityTestAccountingLabel = "test USDC, not real, not redeemable"

// CapabilityProblemTuple is one exact canonical public error contract.
type CapabilityProblemTuple struct {
	Code         string
	Status       int
	Retry        string
	RecoveryHint string
}

// CapabilityProblemTuples contains every exact public error contract.
var CapabilityProblemTuples = [...]CapabilityProblemTuple{
	{Code: "allowance_exhausted", Status: 429, Retry: "state_change", RecoveryHint: "wait for the next subscription period or upgrade"},
	{Code: "allowance_exhausted", Status: 503, Retry: "state_change", RecoveryHint: "capability invocation is disabled"},
	{Code: "authentication_required", Status: 401, Retry: "never", RecoveryHint: "send an active agent-scoped test credential"},
	{Code: "authority_changed", Status: 409, Retry: "state_change", RecoveryHint: "reload the binding authority before retrying"},
	{Code: "binding_not_approved", Status: 403, Retry: "state_change", RecoveryHint: "ask an authorized human to approve or resume the binding"},
	{Code: "binding_not_found", Status: 404, Retry: "state_change", RecoveryHint: "create and approve a pinned binding"},
	{Code: "binding_not_found", Status: 404, Retry: "state_change", RecoveryHint: "use a published capability endpoint"},
	{Code: "budget_exceeded", Status: 402, Retry: "state_change", RecoveryHint: "fund the agent's test USDC balance"},
	{Code: "budget_exceeded", Status: 403, Retry: "state_change", RecoveryHint: "increase the approved binding budget"},
	{Code: "budget_exceeded", Status: 429, Retry: "state_change", RecoveryHint: "wait for the binding velocity window"},
	{Code: "capability_suspended", Status: 409, Retry: "state_change", RecoveryHint: "wait for the provider or capability to be restored"},
	{Code: "credential_wrong_mode", Status: 403, Retry: "never", RecoveryHint: "use an agent-scoped test credential"},
	{Code: "credential_wrong_mode", Status: 403, Retry: "state_change", RecoveryHint: "use an active agent-scoped test credential"},
	{Code: "execution_configuration_invalid", Status: 503, Retry: "state_change", RecoveryHint: "hosted execution is temporarily unavailable"},
	{Code: "execution_provider_unavailable", Status: 502, Retry: "never", RecoveryHint: "contact the capability provider"},
	{Code: "execution_provider_unavailable", Status: 503, Retry: "reconcile", RecoveryHint: "check the invocation status before retrying"},
	{Code: "execution_provider_unavailable", Status: 503, Retry: "same_key", RecoveryHint: "retry with the same idempotency key"},
	{Code: "execution_provider_unavailable", Status: 503, Retry: "same_key", RecoveryHint: "retry with the same key after the provider recovers"},
	{Code: "execution_timeout", Status: 504, Retry: "same_key", RecoveryHint: "retry with the same key after checking status"},
	{Code: "idempotency_conflict", Status: 400, Retry: "never", RecoveryHint: "send one valid Idempotency-Key per logical invocation"},
	{Code: "idempotency_conflict", Status: 409, Retry: "never", RecoveryHint: "reuse the key only with the original request"},
	{Code: "input_schema_invalid", Status: 400, Retry: "never", RecoveryHint: "reduce the invocation input"},
	{Code: "input_schema_invalid", Status: 400, Retry: "never", RecoveryHint: "send bounded JSON matching the published input schema"},
	{Code: "input_schema_invalid", Status: 400, Retry: "never", RecoveryHint: "send valid JSON matching the published schema"},
	{Code: "input_schema_invalid", Status: 400, Retry: "never", RecoveryHint: "send input matching the published schema"},
	{Code: "internal_error", Status: 500, Retry: "never", RecoveryHint: "contact support with the correlation ID"},
	{Code: "invocation_expired", Status: 410, Retry: "never", RecoveryHint: "start a new invocation with a new idempotency key"},
	{Code: "output_schema_invalid", Status: 502, Retry: "never", RecoveryHint: "contact the capability provider"},
	{Code: "permission_denied", Status: 403, Retry: "never", RecoveryHint: "the approved version does not allow this transport"},
	{Code: "permission_denied", Status: 403, Retry: "state_change", RecoveryHint: "human approval is required for this binding"},
	{Code: "permission_denied", Status: 403, Retry: "state_change", RecoveryHint: "the agent policy denies this capability invocation"},
	{Code: "permission_denied", Status: 403, Retry: "state_change", RecoveryHint: "the agent policy requires approval for this invocation"},
	{Code: "version_mismatch", Status: 409, Retry: "state_change", RecoveryHint: "the pinned version is unavailable"},
}
