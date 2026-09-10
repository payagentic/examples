package payagentic

import (
	"math"
	"math/big"
	"strconv"
)

// RFC 8785's finite binary64 spellings need at most 25 bytes. Even the full
// fixed exact decimal expansion of negative minimum subnormal needs only 1,077
// bytes: sign + "0." + 323 leading zeroes + the 751 digits of 5^1074. A 2 KiB
// ceiling therefore covers canonical and exact binary64 decimals, including
// the regression corpus, while bounding lexical and arbitrary-precision work.
const maxBinary64JSONNumberBytes = 2 * 1024

type binary64JSONNumberDecision uint8

const (
	binary64JSONNumberRejected binary64JSONNumberDecision = iota
	binary64JSONNumberNeedsExactComparison
	binary64JSONNumberValidZero
)

func preflightBinary64JSONNumber(text string) (float64, binary64JSONNumberDecision) {
	if len(text) > maxBinary64JSONNumberBytes {
		return 0, binary64JSONNumberRejected
	}

	// The lexical limit bounds exponent digit scanning. ParseFloat validates
	// syntax and overflow without exposing an integer exponent to this code.
	binary64, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(binary64) || math.IsInf(binary64, 0) {
		return 0, binary64JSONNumberRejected
	}
	if binary64 != 0 {
		return binary64, binary64JSONNumberNeedsExactComparison
	}

	// ParseFloat reports deep underflow as zero without an error. Scan only
	// mantissa digits: exponent digits do not make mathematical zero nonzero.
	for index := 0; index < len(text); index++ {
		character := text[index]
		if character == 'e' || character == 'E' {
			return binary64, binary64JSONNumberValidZero
		}
		if character >= '1' && character <= '9' {
			return 0, binary64JSONNumberRejected
		}
	}
	return binary64, binary64JSONNumberValidZero
}

func parseBoundedBinary64JSONNumberRat(text string) (*big.Rat, bool) {
	if len(text) > maxBinary64JSONNumberBytes {
		return nil, false
	}
	return new(big.Rat).SetString(text)
}
