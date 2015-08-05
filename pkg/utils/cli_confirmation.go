package utils

import (
	"fmt"
	"io"
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

func AskForConfirmation(w io.Writer, r io.Reader, prompt string) (bool, error) {
	fmt.Fprint(w, prompt)

	var input string
	if _, err := fmt.Fscanln(r, &input); err != nil {
		return false, err
	}

	input = strings.ToLower(input)

	if _, ok := yesResponses[input]; ok {
		return true, nil
	} else if _, ok := noResponses[input]; ok {
		return false, nil
	} else {
		return AskForConfirmation(w, r, prompt)
	}
}
