package cli

import (
	"fmt"
)

func writeError(err error) {
	fmt.Printf("Encountered a fatal error:\n\t%v\n", err)
}
