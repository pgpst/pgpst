package smtpd

import (
	"fmt"
	"strings"
)

// Gets the address from a string
func parseAddress(input string) (string, error) {
	// Trim it from spaces
	input = strings.TrimSpace(input)

	// Minimal length must be 3 or more
	if len(input) < 3 {
		return "", fmt.Errorf("Ill-formatted e-mail address: %s", input)
	}

	// Ensure that the string starts and ends with gt and lt
	if input[0] != '<' || input[len(input)-1] != '>' {
		return "", fmt.Errorf("Ill-formatted e-mail address: %s", input)
	}

	// It must contain an at sign
	if strings.Count(input, "@") != 1 {
		return "", fmt.Errorf("Ill-formatted e-mail address: %s", input)
	}

	// Return the parsed email
	return input[1 : len(input)-1], nil
}
