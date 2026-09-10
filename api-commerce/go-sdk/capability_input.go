package payagentic

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
)

const maxCapabilityInputDepth = 1_000

var errCapabilityInputJSONNumber = errors.New(
	"capability-input: JSON number is invalid or not exactly representable in binary64",
)

// canonicalCapabilityInputBytes validates and canonicalizes the deliberately
// restricted CapabilityInput domain without invoking caller-owned marshalers.
func canonicalCapabilityInputBytes(input CapabilityInput) ([]byte, error) {
	normalized, err := normalizeCapabilityInputValue(input, 0)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("capability-input: encode normalized JSON: %w", err)
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, fmt.Errorf("capability-input: canonicalize normalized JSON: %w", err)
	}
	return canonical, nil
}

func normalizeCapabilityInputValue(value any, depth int) (any, error) {
	if depth > maxCapabilityInputDepth {
		return nil, fmt.Errorf("capability-input: JSON nesting exceeds %d levels", maxCapabilityInputDepth)
	}

	switch value := value.(type) {
	case nil, bool:
		return value, nil
	case string:
		if !utf8.ValidString(value) {
			return nil, fmt.Errorf("capability-input: invalid Unicode string")
		}
		return value, nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("capability-input: non-finite number is not representable")
		}
		return value, nil
	case json.Number:
		return normalizeCapabilityExactJSONNumber(value)
	case int:
		return normalizeCapabilitySignedInteger(int64(value))
	case int8:
		return normalizeCapabilitySignedInteger(int64(value))
	case int16:
		return normalizeCapabilitySignedInteger(int64(value))
	case int32:
		return normalizeCapabilitySignedInteger(int64(value))
	case int64:
		return normalizeCapabilitySignedInteger(value)
	case uint:
		return normalizeCapabilityUnsignedInteger(uint64(value))
	case uint8:
		return normalizeCapabilityUnsignedInteger(uint64(value))
	case uint16:
		return normalizeCapabilityUnsignedInteger(uint64(value))
	case uint32:
		return normalizeCapabilityUnsignedInteger(uint64(value))
	case uint64:
		return normalizeCapabilityUnsignedInteger(value)
	case uintptr:
		return normalizeCapabilityUnsignedInteger(uint64(value))
	case []any:
		if value == nil {
			return nil, nil
		}
		copy := make([]any, len(value))
		for index, item := range value {
			normalized, err := normalizeCapabilityInputValue(item, depth+1)
			if err != nil {
				return nil, err
			}
			copy[index] = normalized
		}
		return copy, nil
	case CapabilityInput:
		return normalizeCapabilityInputObject(map[string]any(value), depth)
	case map[string]any:
		return normalizeCapabilityInputObject(value, depth)
	default:
		return nil, fmt.Errorf(
			"capability-input: unsupported Go value type %T; use the explicit capability JSON value domain",
			value,
		)
	}
}

func normalizeCapabilityInputObject(value map[string]any, depth int) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	copy := make(map[string]any, len(value))
	for key, item := range value {
		if !utf8.ValidString(key) {
			return nil, fmt.Errorf("capability-input: invalid Unicode object key")
		}
		normalized, err := normalizeCapabilityInputValue(item, depth+1)
		if err != nil {
			return nil, err
		}
		copy[key] = normalized
	}
	return copy, nil
}

func normalizeCapabilityExactJSONNumber(number json.Number) (float64, error) {
	text := number.String()
	if text != strings.TrimSpace(text) || !json.Valid([]byte(text)) {
		return 0, errCapabilityInputJSONNumber
	}
	binary64, decision := preflightBinary64JSONNumber(text)
	switch decision {
	case binary64JSONNumberRejected:
		return 0, errCapabilityInputJSONNumber
	case binary64JSONNumberValidZero:
		return binary64, nil
	case binary64JSONNumberNeedsExactComparison:
		// Continue below with the stricter capability-source comparison.
	default:
		return 0, errCapabilityInputJSONNumber
	}
	source, ok := parseBoundedBinary64JSONNumberRat(text)
	if !ok {
		return 0, errCapabilityInputJSONNumber
	}
	exactBinary64 := new(big.Rat).SetFloat64(binary64)
	if exactBinary64 == nil || source.Cmp(exactBinary64) != 0 {
		return 0, errCapabilityInputJSONNumber
	}
	return binary64, nil
}

func normalizeCapabilitySignedInteger(value int64) (float64, error) {
	binary64 := float64(value)
	source := new(big.Rat).SetInt64(value)
	if source.Cmp(new(big.Rat).SetFloat64(binary64)) != 0 {
		return 0, fmt.Errorf("capability-input: integer would lose precision in binary64")
	}
	return binary64, nil
}

func normalizeCapabilityUnsignedInteger(value uint64) (float64, error) {
	binary64 := float64(value)
	source := new(big.Rat).SetUint64(value)
	if math.IsInf(binary64, 0) || source.Cmp(new(big.Rat).SetFloat64(binary64)) != 0 {
		return 0, fmt.Errorf("capability-input: integer would lose precision in binary64")
	}
	return binary64, nil
}
