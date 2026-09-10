package payagentic

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SessionKey represents a time-scoped signing key for request authentication.
// Session keys allow agents to sign requests without exposing the primary API key.
type SessionKey struct {
	// KeyID is the unique identifier for this session key.
	KeyID string `json:"key_id"`
	// Secret is the HMAC secret used for signing.
	Secret string `json:"secret"`
	// ExpiresAt is when the session key expires.
	ExpiresAt time.Time `json:"expires_at"`
	// Scopes lists the permitted API scopes for this key.
	Scopes []string `json:"scopes,omitempty"`
}

// IsExpired reports whether the session key has expired.
func (sk *SessionKey) IsExpired() bool {
	return time.Now().After(sk.ExpiresAt)
}

// Sign produces an HMAC-SHA256 signature for the given message using the session key secret.
func (sk *SessionKey) Sign(message string) (string, error) {
	if sk.IsExpired() {
		return "", fmt.Errorf("session key %s has expired", sk.KeyID)
	}
	mac := hmac.New(sha256.New, []byte(sk.Secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// SignRequest produces a signature string suitable for the Authorization header.
// The canonical message is: METHOD\nPATH\nTIMESTAMP
func (sk *SessionKey) SignRequest(method, path string, ts time.Time) (string, error) {
	canonical := method + "\n" + path + "\n" + ts.UTC().Format(time.RFC3339)
	sig, err := sk.Sign(canonical)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("SessionKey %s:%s:%s", sk.KeyID, ts.UTC().Format(time.RFC3339), sig), nil
}
