package cli

import (
	"io"

	"github.com/pgpst/pgpst/internal/github.com/pzduniak/cli"
)

const Reader = 0

func Run(r io.Reader, w io.Writer, args []string) (int, error) {
	app := cli.NewApp()

	app.Name = "pgpst-cli"
	app.Usage = "manages pgp.st installation"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "rethinkdb",
			Value:  "rethinkdb://127.0.0.1:28015/prod",
			Usage:  "rethinkdb connection URI",
			EnvVar: "RETHINKDB",
		},
		cli.StringFlag{
			Name:   "default_domain",
			Value:  "pgp.st",
			Usage:  "default domain to use in created data",
			EnvVar: "DEFAULT_DOMAIN",
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
						cli.BoolFlag{
							Name:  "dry",
							Usage: "Start a dry run",
						},
					},
					Action: accountsAdd,
				},
				{
					Name:  "list",
					Usage: "lists accounts",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Output JSON",
						},
					},
					Action: accountsList,
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
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Read JSON from stdin",
						},
					},
					Action: addressesAdd,
				},
				{
					Name:  "list",
					Usage: "lists addresses",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Output JSON",
						},
					},
					Action: addressesList,
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
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Read JSON from stdin",
						},
					},
					Action: applicationsAdd,
				},
				{
					Name:  "list",
					Usage: "lists applications",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Output JSON",
						},
					},
					Action: applicationsList,
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
						cli.BoolFlag{
							Name:  "dry",
							Usage: "Start a dry run",
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
		{
			Name:    "tokens",
			Aliases: []string{"toks"},
			Usage:   "Manage auth and activate tokens",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "creates a new token",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Read JSON from stdin",
						},
					},
					Action: tokensAdd,
				},
				{
					Name:  "list",
					Usage: "lists tokens",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "json",
							Usage: "Output JSON",
						},
					},
					Action: tokensList,
				},
			},
		},
	}

	app.Writer = w
	app.Env["reader"] = r

	return app.Run(args)
}
