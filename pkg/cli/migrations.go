package cli

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"

	"github.com/pgpst/pgpst/pkg/utils"
)

type migration struct {
	Revision int
	Name     string
	Migrate  func(r.ConnectOpts, *r.Session) error
	Revert   func(r.ConnectOpts, *r.Session) error
}

var migrations = []migration{
	{
		Revision: 0,
		Name:     "create_tables",
		Migrate: func(opts r.ConnectOpts, session *r.Session) error {
			if err := utils.MultiExec(
				session,
				r.DB(opts.Database).TableCreate("accounts"),
				r.Table("accounts").IndexCreate("alt_email"),
				r.Table("accounts").IndexCreate("main_address"),
				r.DB(opts.Database).TableCreate("addresses"),
				r.DB(opts.Database).TableCreate("applications"),
				r.DB(opts.Database).TableCreate("emails"),
				r.DB(opts.Database).TableCreate("keys"),
				r.DB(opts.Database).TableCreate("labels"),
				r.DB(opts.Database).TableCreate("resources"),
				r.DB(opts.Database).TableCreate("scope"),
				r.DB(opts.Database).TableCreate("thread"),
				r.DB(opts.Database).TableCreate("token"),
			); err != nil {
				return err
			}

			return nil
		},
		Revert: func(opts r.ConnectOpts, session *r.Session) error {
			if err := utils.MultiExec(
				session,
				r.DB(opts.Database).TableDrop("accounts"),
				r.DB(opts.Database).TableDrop("addresses"),
				r.DB(opts.Database).TableDrop("applications"),
				r.DB(opts.Database).TableDrop("emails"),
				r.DB(opts.Database).TableDrop("keys"),
				r.DB(opts.Database).TableDrop("labels"),
				r.DB(opts.Database).TableDrop("resources"),
				r.DB(opts.Database).TableDrop("scope"),
				r.DB(opts.Database).TableDrop("thread"),
				r.DB(opts.Database).TableDrop("token"),
			); err != nil {
				return err
			}

			return nil
		},
	},
}
