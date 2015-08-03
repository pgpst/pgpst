package cli

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
)

type migration struct {
	Revision int
	Name     string
	Migrate  func(r.ConnectOpts) []r.Term
	Revert   func(r.ConnectOpts) []r.Term
}

var migrations = []migration{
	{
		Revision: 0,
		Name:     "create_tables",
		Migrate: func(opts r.ConnectOpts) []r.Term {
			return []r.Term{
				r.DB(opts.Database).TableCreate("accounts"),
				r.Table("accounts").IndexCreate("alt_email"),
				r.Table("accounts").IndexCreate("main_address"),
				r.DB(opts.Database).TableCreate("addresses"),
				r.DB(opts.Database).TableCreate("applications"),
				r.DB(opts.Database).TableCreate("emails"),
				r.DB(opts.Database).TableCreate("keys"),
				r.DB(opts.Database).TableCreate("labels"),
				r.DB(opts.Database).TableCreate("resources"),
				r.DB(opts.Database).TableCreate("thread"),
				r.DB(opts.Database).TableCreate("tokens"),
			}
		},
		Revert: func(opts r.ConnectOpts) []r.Term {
			return []r.Term{
				r.DB(opts.Database).TableDrop("accounts"),
				r.DB(opts.Database).TableDrop("addresses"),
				r.DB(opts.Database).TableDrop("applications"),
				r.DB(opts.Database).TableDrop("emails"),
				r.DB(opts.Database).TableDrop("keys"),
				r.DB(opts.Database).TableDrop("labels"),
				r.DB(opts.Database).TableDrop("resources"),
				r.DB(opts.Database).TableDrop("thread"),
				r.DB(opts.Database).TableDrop("tokens"),
			}
		},
	},
}
