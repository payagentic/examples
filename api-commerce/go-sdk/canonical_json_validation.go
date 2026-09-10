package payagentic

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
)

var (
	errGenericCanonicalInvalidJSON  = errors.New("emitted bytes are not exactly one valid JSON value")
	errGenericCanonicalInvalidUTF8  = errors.New("emitted JSON is not valid UTF-8")
	errGenericCanonicalInvalidIJSON = errors.New("emitted JSON contains invalid Unicode data")
	errGenericCanonicalNumberDomain = errors.New("emitted JSON number is outside the lossless binary64 domain")
)

// validateGenericCanonicalJSON validates only the bytes emitted by the single
// encoding/json marshal. It must not inspect the source graph or invoke any
// caller-owned marshaler again.
func validateGenericCanonicalJSON(raw []byte) error {
	if !utf8.Valid(raw) {
		return errGenericCanonicalInvalidUTF8
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var emitted any
	if err := decoder.Decode(&emitted); err != nil {
		return errGenericCanonicalInvalidJSON
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errGenericCanonicalInvalidJSON
	}

	// encoding/json accepts malformed UTF-16 escape sequences by replacing
	// them with U+FFFD. RFC 8785 requires such non-Unicode input to fail, while
	// a genuinely encoded U+FFFD remains valid and distinguishable here.
	if err := validateGenericCanonicalJSONStringEscapes(raw); err != nil {
		return err
	}
	return validateGenericCanonicalJSONValue(emitted)
}

func validateGenericCanonicalJSONValue(value any) error {
	switch value := value.(type) {
	case nil, bool, string:
		return nil
	case json.Number:
		return validateGenericCanonicalJSONNumber(value)
	case []any:
		for _, item := range value {
			if err := validateGenericCanonicalJSONValue(item); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		for _, item := range value {
			if err := validateGenericCanonicalJSONValue(item); err != nil {
				return err
			}
		}
		return nil
	default:
		return errGenericCanonicalInvalidJSON
	}
}

func validateGenericCanonicalJSONNumber(number json.Number) error {
	text := number.String()
	binary64, decision := preflightBinary64JSONNumber(text)
	switch decision {
	case binary64JSONNumberRejected:
		return errGenericCanonicalNumberDomain
	case binary64JSONNumberValidZero:
		return nil
	case binary64JSONNumberNeedsExactComparison:
		return validateGenericCanonicalJSONNumberExact(text, binary64)
	default:
		return errGenericCanonicalNumberDomain
	}
}

func validateGenericCanonicalJSONNumberExact(text string, binary64 float64) error {
	source, ok := parseBoundedBinary64JSONNumberRat(text)
	if !ok {
		return errGenericCanonicalInvalidJSON
	}

	exactBinary64 := new(big.Rat).SetFloat64(binary64)
	if exactBinary64 == nil {
		return errGenericCanonicalNumberDomain
	}

	// Integral claims, including decimal- or exponent-form spellings such as
	// 1.0 and 1e0, must retain their exact mathematical value in binary64.
	if source.IsInt() {
		if source.Cmp(exactBinary64) != 0 {
			return errGenericCanonicalNumberDomain
		}
		return nil
	}

	shortest, err := jcs.NumberToJSON(binary64)
	if err != nil {
		return errGenericCanonicalNumberDomain
	}
	shortestValue, ok := parseBoundedBinary64JSONNumberRat(shortest)
	if !ok {
		return errGenericCanonicalNumberDomain
	}

	// A nonintegral source is unambiguous when it is either the shortest JCS
	// spelling of the binary64 value (for example 0.1 or 5e-324) or the exact
	// decimal expansion of that binary64 value.
	if source.Cmp(shortestValue) == 0 || source.Cmp(exactBinary64) == 0 {
		return nil
	}
	return errGenericCanonicalNumberDomain
}

func validateGenericCanonicalJSONStringEscapes(raw []byte) error {
	inString := false
	for index := 0; index < len(raw); index++ {
		switch character := raw[index]; {
		case !inString:
			if character == '"' {
				inString = true
			}
		case character == '"':
			inString = false
		case character == '\\':
			index++
			if index >= len(raw) {
				return errGenericCanonicalInvalidJSON
			}
			if raw[index] != 'u' {
				continue
			}
			if index+4 >= len(raw) {
				return errGenericCanonicalInvalidJSON
			}
			codeUnit, ok := parseGenericCanonicalJSONHexQuad(raw[index+1 : index+5])
			if !ok {
				return errGenericCanonicalInvalidJSON
			}
			index += 4

			switch {
			case codeUnit >= 0xd800 && codeUnit <= 0xdbff:
				if index+6 >= len(raw) || raw[index+1] != '\\' || raw[index+2] != 'u' {
					return errGenericCanonicalInvalidIJSON
				}
				low, ok := parseGenericCanonicalJSONHexQuad(raw[index+3 : index+7])
				if !ok || low < 0xdc00 || low > 0xdfff {
					return errGenericCanonicalInvalidIJSON
				}
				index += 6
			case codeUnit >= 0xdc00 && codeUnit <= 0xdfff:
				return errGenericCanonicalInvalidIJSON
			}
		}
	}
	return nil
}

func parseGenericCanonicalJSONHexQuad(raw []byte) (uint16, bool) {
	if len(raw) != 4 {
		return 0, false
	}
	var value uint16
	for _, digit := range raw {
		value *= 16
		switch {
		case digit >= '0' && digit <= '9':
			value += uint16(digit - '0')
		case digit >= 'a' && digit <= 'f':
			value += uint16(digit-'a') + 10
		case digit >= 'A' && digit <= 'F':
			value += uint16(digit-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
