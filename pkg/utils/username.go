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
