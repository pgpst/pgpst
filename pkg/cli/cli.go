package cli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func Run() {
	app := cli.NewApp()

	app.Name = "pgpst-cli"
	app.Usage = "manages pgp.st installation"

	app.Commands = []cli.Command{
		{
			Name:    "accounts",
			Aliases: []string{"accs"},
			Usage:   "manage accounts",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "creates a new account",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
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
			Name:    "applications",
			Aliases: []string{"apps"},
			Usage:   "manage applications",
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
			Usage:   "database operations",
			Subcommands: []cli.Command{
				{
					Name:  "migrate",
					Usage: "run all migrations",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
				},
				{
					Name:  "version",
					Usage: "compare database version to the tool's",
					Action: func(c *cli.Context) {
						fmt.Printf("%+v\n", c.Args().Tail())
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
