package main

import (
	"os"

	"github.com/pgpst/pgpst/pkg/cli"
	"github.com/pgpst/pgpst/pkg/utils"
)

func main() {
	defer utils.CLIRecovery(os.Stderr)
	cli.Run(os.Stdin, os.Stdout, os.Args)
}
