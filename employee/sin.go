package employee

import (
	"strings"
	"unicode"
)

// NormalizeSIN strips non-digits; returns empty if not exactly 9 digits.
func NormalizeSIN(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) != 9 {
		return ""
	}
	return out
}

// ValidateSIN checks Canadian SIN format: 9 digits, not all zeros, passes Luhn check.
func ValidateSIN(s string) bool {
	n := NormalizeSIN(s)
	if n == "" {
		return false
	}
	if n == "000000000" {
		return false
	}
	return luhnValid(n)
}

func luhnValid(digits string) bool {
	sum := 0
	alt := false
	for i := len(digits) - 1; i >= 0; i-- {
		n := int(digits[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}
