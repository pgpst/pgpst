package cli

import (
	"fmt"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/cli"

	"github.com/pgpst/pgpst/pkg/utils"
)

func writeError(err error) {
	fmt.Printf("Encountered a fatal error:\n\t%v\n", err)
}

func connectToRethinkDB(c *cli.Context) (*r.ConnectOpts, *r.Session, bool) {
	opts, err := utils.ParseRethinkDBString(c.GlobalString("rethinkdb"))
	if err != nil {
		writeError(err)
		return nil, nil, false
	}
	session, err := r.Connect(opts)
	if err != nil {
		writeError(err)
		return nil, nil, false
	}

	return &opts, session, true
}
