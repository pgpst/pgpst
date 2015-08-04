package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// Username normalization
var rNotASCII = regexp.MustCompile(`[^\w\.]`)

func NormalizeUsername(input string) string {
	return rNotASCII.ReplaceAllString(
		strings.ToLowerSpecial(unicode.TurkishCase, input),
		"",
	)
}

func NormalizeAddress(input string) string {
	parts := strings.SplitN(input, "@", 2)

	return NormalizeUsername(parts[0]) + "@" + strings.ToLowerSpecial(unicode.TurkishCase, parts[1])
}
