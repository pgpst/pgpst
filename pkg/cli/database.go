package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"

	"github.com/pgpst/pgpst/pkg/utils"
)

func getDatabaseVersion(opts r.ConnectOpts, session *r.Session) (int, error) {
	cursor, err := r.Branch(
		r.DB(opts.Database).TableList().Contains("migration_status"),
		r.DB(opts.Database).Table("migration_status").Get("revision").Field("value"),
		r.DB(opts.Database).TableCreate("migration_status").Do(func() r.Term {
			return r.DB(opts.Database).Table("migration_status").Insert(map[string]interface{}{
				"id":    "revision",
				"value": -1,
			}).Do(func() int {
				return -1
			})
		}),
	).Run(session)
	if err != nil {
		return -1, err
	}
	defer cursor.Close()
	var result int
	if err := cursor.One(&result); err != nil {
		return -1, err
	}

	return result, nil
}

func databaseMigrate(c *cli.Context) {
	opts, err := utils.ParseRethinkDBString(c.GlobalString("rethinkdb"))
	if err != nil {
		writeError(err)
		return
	}

	session, err := r.Connect(opts)
	if err != nil {
		writeError(err)
		return
	}

	version, err := getDatabaseVersion(opts, session)
	if err != nil {
		writeError(err)
		return
	}

	fmt.Printf("Current database schema's version is %d.\n", version)
	fmt.Printf("Latest migration's version is %d.\n", len(migrations)-1)
}
