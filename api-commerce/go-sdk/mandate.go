package payagentic

// Mandate envelope, signer, and presentation helpers (M6).
//
// Mirrors the TypeScript/Python SDKs and the Rust payagentic-mandate crate
// byte-for-byte. The JWS protected header is
// {"alg":"EdDSA","kid":<kid>,"typ":"JWT"} and the payload is encoded with
// RFC 8785 canonical JSON before signing so signatures verify identically
// across SDKs.

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	MandateSupportedVersion = 1
	MandateSupportedAlg     = "EdDSA"
	DefaultMandateAudience  = "payagentic-mandate"
)

// MandateSigner produces an Ed25519 signature over a mandate signing input.
//
// Kid is echoed in the JWS protected header so the verifier can look up
// the matching public key.
type MandateSigner interface {
	Kid() string
	Sign(message []byte) ([]byte, error)
}

// LocalMandateSigner holds an Ed25519 private key in memory and signs
// locally. Use for development, tests, and short-lived agent processes.
// Production operator signers should wrap Vault or an HSM.
type LocalMandateSigner struct {
	kid string
	sk  ed25519.PrivateKey
}

// NewLocalMandateSigner constructs a signer from a 32-byte seed (the
// canonical Ed25519 "private key" representation matching the TS and
// Python SDKs).
func NewLocalMandateSigner(kid string, seed []byte) (*LocalMandateSigner, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf(
			"LocalMandateSigner: Ed25519 seed must be %d bytes, got %d",
			ed25519.SeedSize, len(seed),
		)
	}
	return &LocalMandateSigner{kid: kid, sk: ed25519.NewKeyFromSeed(seed)}, nil
}

// GenerateLocalMandateSigner produces a new random Ed25519 key.
func GenerateLocalMandateSigner(kid string) (*LocalMandateSigner, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("ed25519 seed gen: %w", err)
	}
	return NewLocalMandateSigner(kid, seed)
}

// Kid implements MandateSigner.
func (s *LocalMandateSigner) Kid() string { return s.kid }

// Sign implements MandateSigner.
func (s *LocalMandateSigner) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(s.sk, message), nil
}

// PublicKey returns the 32-byte Ed25519 public key.
func (s *LocalMandateSigner) PublicKey() ed25519.PublicKey {
	return s.sk.Public().(ed25519.PublicKey)
}

// ─────────────────────────────────────────────────────────────────────────
// JWS
// ─────────────────────────────────────────────────────────────────────────

// MandateMoney is the JSON shape for Money used in mandate payloads.
// Currency defaults to USDC in v1.
type MandateMoney struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// PresentationClaim is the per-charge claim asserted by the agent at
// presentation time.
type PresentationClaim struct {
	Amount      MandateMoney `json:"amount"`
	MerchantID  string       `json:"merchant_id"`
	MCC         string       `json:"mcc"`
	OccurredAt  int64        `json:"occurred_at"`
	Geo         string       `json:"geo"`
	ResourceURL string       `json:"resource_url"`
}

// BuildPresentationParams configures PresentMandate.
type BuildPresentationParams struct {
	AgentSpiffeID string
	Audience      string
	Grants        []string
	Claim         PresentationClaim
	IAT           int64 // optional; defaults to time.Now().Unix()
	Exp           int64 // optional; defaults to IAT + 300
	JTI           string
}

func b64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// SignJWS signs payload as a compact-form JWS with alg=EdDSA.
func SignJWS(payload any, signer MandateSigner) (string, error) {
	header := map[string]any{
		"alg": MandateSupportedAlg,
		"kid": signer.Kid(),
		"typ": "JWT",
	}
	hb, err := CanonicalJSONBytes(header)
	if err != nil {
		return "", fmt.Errorf("canonical header: %w", err)
	}
	pb, err := CanonicalJSONBytes(payload)
	if err != nil {
		return "", fmt.Errorf("canonical payload: %w", err)
	}
	signingInput := b64url(hb) + "." + b64url(pb)
	sig, err := signer.Sign([]byte(signingInput))
	if err != nil {
		return "", fmt.Errorf("signer: %w", err)
	}
	return signingInput + "." + b64url(sig), nil
}

func newJTI() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Extremely unlikely; fall back to a deterministic-ish value seeded
		// by current time so we never panic in the SDK hot path.
		sum := sha256.Sum256(fmt.Appendf(nil, "%d", time.Now().UnixNano()))
		copy(b[:], sum[:16])
	}
	return hex.EncodeToString(b[:])
}

// PresentMandate builds and signs a presentation JWS for the given claim.
func PresentMandate(params BuildPresentationParams, signer MandateSigner) (string, error) {
	if signer == nil {
		return "", errors.New("PresentMandate: signer is required")
	}
	if params.AgentSpiffeID == "" {
		return "", errors.New("PresentMandate: AgentSpiffeID is required")
	}
	audience := params.Audience
	if audience == "" {
		audience = DefaultMandateAudience
	}
	iat := params.IAT
	if iat == 0 {
		iat = time.Now().Unix()
	}
	exp := params.Exp
	if exp == 0 {
		exp = iat + 300
	}
	jti := params.JTI
	if jti == "" {
		jti = newJTI()
	}
	// Marshal claim to map[string]any via JSON round-trip so the canonical
	// encoder sorts its keys and matches the TS / Python output byte-for-byte.
	claimRaw, err := json.Marshal(params.Claim)
	if err != nil {
		return "", fmt.Errorf("marshal claim: %w", err)
	}
	var claimMap map[string]any
	dec := json.NewDecoder(bytes.NewReader(claimRaw))
	dec.UseNumber()
	if err := dec.Decode(&claimMap); err != nil {
		return "", fmt.Errorf("decode claim: %w", err)
	}
	payload := map[string]any{
		"v":      json.Number(fmt.Sprintf("%d", MandateSupportedVersion)),
		"typ":    "presentation",
		"iss":    params.AgentSpiffeID,
		"aud":    audience,
		"jti":    jti,
		"iat":    json.Number(fmt.Sprintf("%d", iat)),
		"exp":    json.Number(fmt.Sprintf("%d", exp)),
		"grants": toAnySlice(params.Grants),
		"claim":  claimMap,
	}
	return SignJWS(payload, signer)
}

func toAnySlice(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────
// Client helpers — issue / revoke
// ─────────────────────────────────────────────────────────────────────────

// IssueGrantParams configures Client.IssueGrant.
type IssueGrantParams struct {
	SubjectAgentID string         `json:"subject_agent_id"`
	Scope          map[string]any `json:"scope"`
	Exp            int64          `json:"exp"`
	NBF            *int64         `json:"nbf,omitempty"`
	MaxDepth       int            `json:"max_depth"`
	ParentJTI      string         `json:"parent_jti,omitempty"`
	IssuerKID      string         `json:"issuer_kid"`
}

// IssuedGrant is the response from POST /v1/mandates/issue.
type IssuedGrant struct {
	JTI         string `json:"jti"`
	EnvelopeJWS string `json:"envelope_jws"`
}

// RevokedGrant is the response from POST /v1/mandates/{jti}/revoke.
type RevokedGrant struct {
	JTI       string `json:"jti"`
	RevokedAt string `json:"revoked_at"`
}

// MandatesService exposes operator-side issue / revoke endpoints.
//
// Construct via Client.Mandates() — it is intentionally not auto-attached
// to NewClient so existing call sites keep working.
type MandatesService struct {
	client *Client
}

// Mandates returns a MandatesService bound to this client.
func (c *Client) Mandates() *MandatesService {
	return &MandatesService{client: c}
}

// Issue calls POST /v1/mandates/issue.
//
// NOTE: /v1/mandates/issue and /v1/mandates/{jti}/revoke are owned by the
// identity service, not exposed in the gateway spec at api/openapi.json.
// Direct HTTP via client.do() is the only option until those endpoints
// are surfaced.
func (s *MandatesService) Issue(ctx context.Context, params IssueGrantParams) (*IssuedGrant, error) {
	var out IssuedGrant
	if err := s.client.do(ctx, "POST", "/v1/mandates/issue", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Revoke calls POST /v1/mandates/{jti}/revoke. Idempotent.
//
// NOTE: identity-service-owned (see Issue's docstring for rationale).
func (s *MandatesService) Revoke(ctx context.Context, jti, reason string) (*RevokedGrant, error) {
	body := map[string]any{}
	if reason != "" {
		body["reason"] = reason
	}
	var out RevokedGrant
	path := "/v1/mandates/" + jti + "/revoke"
	if err := s.client.do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
