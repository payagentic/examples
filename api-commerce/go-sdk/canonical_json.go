package payagentic

import (
	"encoding/json"
	"fmt"

	"github.com/gowebpki/jcs"
)

// CanonicalJSONBytes returns the RFC 8785 canonical JSON UTF-8 bytes of value.
//
// CanonicalJSONBytes calls encoding/json.Marshal exactly once, honoring its
// ordinary struct fields and tags, typed maps and slices, json.RawMessage, and
// JSON or text marshalers. It validates the emitted UTF-8/I-JSON bytes and
// rejects lossy integral claims before applying the pinned JCS transform;
// shortest RFC 8785 nonintegral decimals remain accepted. Emitted numeric
// tokens are bounded before exact validation. Capability invocation applies
// its stricter exact-source input rules separately.
func CanonicalJSONBytes(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("canonical-json: marshal JSON: %w", err)
	}
	if err := validateGenericCanonicalJSON(raw); err != nil {
		return nil, fmt.Errorf("canonical-json: validate marshaled JSON: %w", err)
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, fmt.Errorf("canonical-json: canonicalize marshaled JSON: %w", err)
	}
	return canonical, nil
}
