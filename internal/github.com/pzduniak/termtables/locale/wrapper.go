// Copyright 2013 Apcera Inc. All rights reserved.

// +build ignore

package main

import (
	"fmt"

	"github.com/pgpst/pgpst/internal/github.com/pzduniak/termtables/locale"
)

func main() {
	fmt.Println(locale.GetCharmap())
}
