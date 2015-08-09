package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pgpst/pgpst/internal/github.com/cheggaaa/pb"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/cli"

	"github.com/pgpst/pgpst/pkg/utils"
)

func getDatabaseVersion(opts *r.ConnectOpts, session *r.Session) (int, error) {
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

func databaseVersion(c *cli.Context) int {
	// Connect to RethinkDB
	opts, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Get the migration status from the database
	version, err := getDatabaseVersion(opts, session)
	if err != nil {
		writeError(c, err)
		return 1
	}

	// Write it to stdout
	fmt.Fprintln(c.App.Writer, version)
	return 0
}

func databaseMigrate(c *cli.Context) int {
	// Connect to RethinkDB
	opts, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Get the migration status from the database
	version, err := getDatabaseVersion(opts, session)
	if err != nil {
		writeError(c, err)
		return 1
	}

	// Show the current migration's status
	fmt.Fprintf(c.App.Writer, "Current database schema's version is %d.\n", version)
	fmt.Fprintf(c.App.Writer, "Latest migration's version is %d.\n", len(migrations)-1)

	// Only proceed if the schema is outdated
	if version >= len(migrations)-1 {
		fmt.Fprintln(c.App.Writer, "Your schema is up to date.")
		return 0
	}

	// I don't know why would anyone use it, but it's here
	if c.Bool("no") {
		fmt.Fprintln(c.App.Writer, "Aborting the command because of the --no option.")
		return 1
	}

	// Ask for confirmations
	if !c.Bool("yes") {
		want, err := utils.AskForConfirmation(
			c.App.Writer,
			c.App.Env["reader"].(io.Reader),
			"Would you like to run "+strconv.Itoa(len(migrations)-1-version)+" migrations? [y/n]: ",
		)
		if err != nil {
			writeError(c, err)
			return 1
		}

		if !want {
			fmt.Fprintln(c.App.Writer, "Aborting the command.")
			return 1
		}
	}

	// Collect all queries
	queries := []r.Term{}
	for _, migration := range migrations[version+1:] {
		queries = append(queries, migration.Migrate(opts)...)
		queries = append(queries, r.Table("migration_status").Get("revision").Update(map[string]interface{}{
			"value": migration.Revision,
		}))
	}

	// Create a new progress bar
	bar := pb.StartNew(len(queries))
	for i, query := range queries {
		if c.Bool("dry") {
			fmt.Fprintf(c.App.Writer, "Executing %s\n", query.String())
		} else {
			if err := query.Exec(session); err != nil {
				bar.FinishPrint("Failed to execute migration #" + strconv.Itoa(i) + ":")
				fmt.Fprintf(c.App.Writer, "\tQuery: %s\n", query.String())
				fmt.Fprintf(c.App.Writer, "\tError: %v\n", err)
				return 1
			}
		}
		bar.Increment()
	}

	// Show a "finished" message
	bar.FinishPrint("Migration completed. " + strconv.Itoa(len(queries)) + " queries executed.")
	return 0
}
