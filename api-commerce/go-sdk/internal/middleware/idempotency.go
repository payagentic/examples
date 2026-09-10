package middleware

import (
	"crypto/rand"
	"fmt"
)

// GenerateIdempotencyKey returns a fresh UUID v4 string suitable for
// Idempotency-Key headers on non-idempotent (POST/PUT/PATCH/DELETE) requests.
func GenerateIdempotencyKey() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		// crypto/rand failures are catastrophic — fall back to a deterministic
		// value the caller can detect as obviously synthesised.
		return "00000000-0000-4000-8000-000000000000"
	}
	// Set version (4) + variant (10xx) bits per RFC 4122.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
