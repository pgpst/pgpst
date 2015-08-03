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

func databaseVersion(c *cli.Context) {
	// Set up a RethinkDB connection
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

	// Get the migration status from the database
	version, err := getDatabaseVersion(opts, session)
	if err != nil {
		writeError(err)
		return
	}

	// Write it to stdout
	fmt.Println(version)
}

func databaseMigrate(c *cli.Context) {
	// Set up a RethinkDB connection
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

	// Get the migration status from the database
	version, err := getDatabaseVersion(opts, session)
	if err != nil {
		writeError(err)
		return
	}

	// Show the current migration's status
	fmt.Printf("Current database schema's version is %d.\n", version)
	fmt.Printf("Latest migration's version is %d.\n", len(migrations)-1)

	// I don't know why would anyone use it, but it's here
	if c.Bool("no") {
		fmt.Println("Aborting the command because of the --no option.")
		return
	}

	// Ask for confirmations
	if !c.Bool("yes") {
		want, err := utils.AskForConfirmation("Would you like to run the %d migrations? [y/n]: ")
		if err != nil {
			writeError(err)
			return
		}

		if !want {
			fmt.Println("Aborting the command.")
		}
	}

	// Run the migrations
}
