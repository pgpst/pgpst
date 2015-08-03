package cli

import (
	"fmt"
	"strconv"

	"github.com/cheggaaa/pb"
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

	// Only proceed if the schema is outdated
	if version >= len(migrations)-1 {
		fmt.Println("Your schema is up to date.")
		return
	}

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
			return
		}
	}

	// Collect all queries
	queries := []r.Term{}
	for _, migration := range migrations[version:] {
		queries = append(queries, migration.Migrate(opts)...)
		queries = append(queries, r.Table("migration_status").Get("revision").Update(map[string]interface{}{
			"value": migration.Revision,
		}))
	}

	// Create a new progress bar
	bar := pb.StartNew(len(queries))
	for i, query := range queries {
		if err := query.Exec(session); err != nil {
			bar.FinishPrint("Failed to execute migration #" + strconv.Itoa(i) + ":")
			fmt.Printf("\tQuery: %s\n", query.String())
			fmt.Printf("\tError: %v\n", err)
			return
		}
		bar.Increment()
	}

	// Show a "finished" message
	bar.FinishPrint("Migration completed. " + strconv.Itoa(len(queries)) + " queries executed.")
}
