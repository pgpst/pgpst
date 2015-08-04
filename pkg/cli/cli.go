package cli

import (
	"fmt"
	"os"

	"github.com/pgpst/pgpst/internal/github.com/codegangsta/cli"
)

func Run() {
	app := cli.NewApp()

	app.Name = "pgpst-cli"
	app.Usage = "manages pgp.st installation"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rethinkdb",
			Value: "rethinkdb://127.0.0.1:28015/prod",
			Usage: "rethinkdb connection URI",
		},
		cli.StringFlag{
			Name:  "default_domain",
			Value: "pgp.st",
			Usage: "default domain to use in created data",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "accounts",
			Aliases: []string{"accs"},
			Usage:   "Manages account objects",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "creates a new account",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Read JSON from stdin",
						},
					},
					Action: accountAdd,
				},
				{
					Name:  "list",
					Usage: "lists accounts",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
				},
			},
		},
		{
			Name:    "addresses",
			Aliases: []string{"addrs"},
			Usage:   "Manages address objects",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "creates a new address",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
				},
				{
					Name:  "list",
					Usage: "lists addresses",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
				},
			},
		},
		{
			Name:    "applications",
			Aliases: []string{"apps"},
			Usage:   "Manage OAuth applications",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "creates a new application",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
				},
				{
					Name:  "list",
					Usage: "lists applications",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
				},
			},
		},
		{
			Name:    "database",
			Aliases: []string{"db"},
			Usage:   "Various database operations",
			Subcommands: []cli.Command{
				{
					Name:  "migrate",
					Usage: "run all migrations",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "yes",
							Usage: "Say yes to the prompt",
						},
						cli.BoolFlag{
							Name:  "no",
							Usage: "Say no to the prompt",
						},
					},
					Action: databaseMigrate,
				},
				{
					Name:   "version",
					Usage:  "compare database version to the tool's",
					Action: databaseVersion,
				},
			},
		},
	}

	app.Run(os.Args)
}
