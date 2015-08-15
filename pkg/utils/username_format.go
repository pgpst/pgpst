package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// Username normalization
var rNotASCII = regexp.MustCompile(`[^\w\.]`)

func RemoveDots(input string) string {
	if strings.Index(input, "@") != -1 {
		parts := strings.SplitN(input, "@", 2)

		return strings.Replace(parts[0], ".", "", -1) + "@" + parts[1]
	}

	return strings.Replace(input, ".", "", -1)
}

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
