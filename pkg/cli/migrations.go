package cli

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
)

type migration struct {
	Revision int
	Name     string
	Migrate  func(*r.ConnectOpts) []r.Term
	Revert   func(*r.ConnectOpts) []r.Term
}

var migrations = []migration{
	{
		Revision: 0,
		Name:     "create_tables",
		Migrate: func(opts *r.ConnectOpts) []r.Term {
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
				r.DB(opts.Database).TableCreate("threads"),
				r.DB(opts.Database).TableCreate("tokens"),
			}
		},
		Revert: func(opts *r.ConnectOpts) []r.Term {
			return []r.Term{
				r.DB(opts.Database).TableDrop("accounts"),
				r.DB(opts.Database).TableDrop("addresses"),
				r.DB(opts.Database).TableDrop("applications"),
				r.DB(opts.Database).TableDrop("emails"),
				r.DB(opts.Database).TableDrop("keys"),
				r.DB(opts.Database).TableDrop("labels"),
				r.DB(opts.Database).TableDrop("resources"),
				r.DB(opts.Database).TableDrop("threads"),
				r.DB(opts.Database).TableDrop("tokens"),
			}
		},
	},
	{
		Revision: 1,
		Name:     "analytic_indexes",
		Migrate: func(opts *r.ConnectOpts) []r.Term {
			return []r.Term{
				r.Table("accounts").IndexCreate("date_created"),
				r.Table("accounts").IndexCreate("date_modified"),

				r.Table("addresses").IndexCreate("owner"),
				r.Table("addresses").IndexCreate("date_created"),
				r.Table("addresses").IndexCreate("date_modified"),

				r.Table("applications").IndexCreate("owner"),
				r.Table("applications").IndexCreate("name"),
				r.Table("applications").IndexCreate("date_created"),
				r.Table("applications").IndexCreate("date_modified"),

				r.Table("emails").IndexCreate("owner"),
				r.Table("emails").IndexCreate("from"),
				r.Table("emails").IndexCreate("to", r.IndexCreateOpts{Multi: true}),
				r.Table("emails").IndexCreate("cc", r.IndexCreateOpts{Multi: true}),
				r.Table("emails").IndexCreate("bcc", r.IndexCreateOpts{Multi: true}),
				r.Table("emails").IndexCreate("files", r.IndexCreateOpts{Multi: true}),
				r.Table("emails").IndexCreate("thread"),
				r.Table("emails").IndexCreate("status"),
				r.Table("emails").IndexCreate("secure"),
				r.Table("emails").IndexCreate("date_created"),
				r.Table("emails").IndexCreate("date_modified"),

				r.Table("keys").IndexCreate("owner"),
				r.Table("keys").IndexCreate("expiry_date"),
				r.Table("keys").IndexCreate("algorithm"),
				r.Table("keys").IndexCreate("key_id"),
				r.Table("keys").IndexCreate("key_id_short"),
				r.Table("keys").IndexCreate("master_key"),
				r.Table("keys").IndexCreate("date_created"),
				r.Table("keys").IndexCreate("date_modified"),

				r.Table("labels").IndexCreate("owner"),
				r.Table("labels").IndexCreate("name"),
				r.Table("labels").IndexCreate("system"),
				r.Table("labels").IndexCreate("date_created"),
				r.Table("labels").IndexCreate("date_modified"),

				r.Table("resources").IndexCreate("owner"),
				r.Table("resources").IndexCreate("tags", r.IndexCreateOpts{Multi: true}),
				r.Table("resources").IndexCreate("date_created"),
				r.Table("resources").IndexCreate("date_modified"),

				r.Table("threads").IndexCreate("owner"),
				r.Table("threads").IndexCreate("emails", r.IndexCreateOpts{Multi: true}),
				r.Table("threads").IndexCreate("labels", r.IndexCreateOpts{Multi: true}),
				r.Table("threads").IndexCreate("members", r.IndexCreateOpts{Multi: true}),
				r.Table("threads").IndexCreate("is_read"),
				r.Table("threads").IndexCreate("is_secure"),
				r.Table("threads").IndexCreate("date_created"),
				r.Table("threads").IndexCreate("date_modified"),

				r.Table("tokens").IndexCreate("owner"),
				r.Table("tokens").IndexCreate("type"),
				r.Table("tokens").IndexCreate("scope", r.IndexCreateOpts{Multi: true}),
				r.Table("tokens").IndexCreate("client_id"),
				r.Table("tokens").IndexCreate("expiry_date"),
				r.Table("tokens").IndexCreate("date_created"),
				r.Table("tokens").IndexCreate("date_modified"),
			}
		},
		Revert: func(opts *r.ConnectOpts) []r.Term {
			return []r.Term{
				r.Table("accounts").IndexDrop("date_created"),
				r.Table("accounts").IndexDrop("date_modified"),

				r.Table("addresses").IndexDrop("owner"),
				r.Table("addresses").IndexDrop("date_created"),
				r.Table("addresses").IndexDrop("date_modified"),

				r.Table("applications").IndexDrop("owner"),
				r.Table("applications").IndexDrop("name"),
				r.Table("applications").IndexDrop("date_created"),
				r.Table("applications").IndexDrop("date_modified"),

				r.Table("emails").IndexDrop("owner"),
				r.Table("emails").IndexDrop("from"),
				r.Table("emails").IndexDrop("to"),
				r.Table("emails").IndexDrop("cc"),
				r.Table("emails").IndexDrop("bcc"),
				r.Table("emails").IndexDrop("files"),
				r.Table("emails").IndexDrop("thread"),
				r.Table("emails").IndexDrop("status"),
				r.Table("emails").IndexDrop("secure"),
				r.Table("emails").IndexDrop("date_created"),
				r.Table("emails").IndexDrop("date_modified"),

				r.Table("keys").IndexDrop("owner"),
				r.Table("keys").IndexDrop("expiry_date"),
				r.Table("keys").IndexDrop("algorithm"),
				r.Table("keys").IndexDrop("key_id"),
				r.Table("keys").IndexDrop("key_id_short"),
				r.Table("keys").IndexDrop("master_key"),
				r.Table("keys").IndexDrop("date_created"),
				r.Table("keys").IndexDrop("date_modified"),

				r.Table("labels").IndexDrop("owner"),
				r.Table("labels").IndexDrop("name"),
				r.Table("labels").IndexDrop("system"),
				r.Table("labels").IndexDrop("date_created"),
				r.Table("labels").IndexDrop("date_modified"),

				r.Table("resources").IndexDrop("owner"),
				r.Table("resources").IndexDrop("tags"),
				r.Table("resources").IndexDrop("date_created"),
				r.Table("resources").IndexDrop("date_modified"),

				r.Table("threads").IndexDrop("owner"),
				r.Table("threads").IndexDrop("emails"),
				r.Table("threads").IndexDrop("labels"),
				r.Table("threads").IndexDrop("members"),
				r.Table("threads").IndexDrop("is_read"),
				r.Table("threads").IndexDrop("is_secure"),
				r.Table("threads").IndexDrop("date_created"),
				r.Table("threads").IndexDrop("date_modified"),

				r.Table("tokens").IndexDrop("owner"),
				r.Table("tokens").IndexDrop("type"),
				r.Table("tokens").IndexDrop("scope"),
				r.Table("tokens").IndexDrop("client_id"),
				r.Table("tokens").IndexDrop("expiry_date"),
				r.Table("tokens").IndexDrop("date_created"),
				r.Table("tokens").IndexDrop("date_modified"),
			}
		},
	},
	{
		Revision: 2,
		Name:     "mailer indices",
		Migrate: func(opts *r.ConnectOpts) []r.Term {
			return []r.Term{
				r.Table("emails").IndexCreateFunc("messageIDOwner", func(row r.Term) []interface{} {
					return []interface{}{
						row.Field("message_id"),
						row.Field("owner"),
					}
				}),
				r.Table("labels").IndexCreateFunc("nameOwnerSystem", func(row r.Term) []interface{} {
					return []interface{}{
						row.Field("name"),
						row.Field("owner"),
						row.Field("system"),
					}
				}),
			}
		},
		Revert: func(opts *r.ConnectOpts) []r.Term {
			return []r.Term{
				r.Table("emails").IndexDrop("messageIDOwner"),
				r.Table("labels").IndexDrop("nameOwnerSystem"),
			}
		},
	},
}
