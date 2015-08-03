package utils

import (
	"fmt"
	"strings"
)

var yesResponses = map[string]struct{}{
	"y":   {},
	"yes": {},
}

var noResponses = map[string]struct{}{
	"n":  {},
	"no": {},
}

func AskForConfirmation(prompt string) (bool, error) {
	fmt.Print(prompt)

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return false, err
	}

	input = strings.ToLower(input)

	if _, ok := yesResponses[input]; ok {
		return true, nil
	} else if _, ok := noResponses[input]; ok {
		return false, nil
	} else {
		return AskForConfirmation(prompt)
	}
}
