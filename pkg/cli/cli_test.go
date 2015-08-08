package cli_test

import (
	"bytes"
	"os"
	"testing"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/cli"
	"github.com/pgpst/pgpst/pkg/utils"
)

func TestCLI(t *testing.T) {
	Convey("All CLI cases should work properly", t, func() {
		// Connect to the server
		opts, err := utils.ParseRethinkDBString(os.Getenv("RETHINKDB"))
		So(err, ShouldBeNil)

		session, err := r.Connect(opts)
		So(err, ShouldBeNil)

		r.DBDrop(opts.Database).Exec(session)
		r.DBCreate(opts.Database).Exec(session)

		// Run database version check
		output := &bytes.Buffer{}
		code, err := cli.Run(os.Stdin, output, []string{
			"pgpst-cli",
			"db",
			"version",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		// Run migration with the --no option
		output.Reset()
		code, err = cli.Run(os.Stdin, output, []string{
			"pgpst-cli",
			"db",
			"migrate",
			"--no",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Run dry migration with stdin not confirmed
		output.Reset()
		input := &bytes.Buffer{}
		input.WriteString("no\n")
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"db",
			"migrate",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Run dry migration with stdin confirmation
		output.Reset()
		input.Reset()
		input.WriteString("yes\n")
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"db",
			"migrate",
			"--dry",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		// Run actual migration and fill the database
		output.Reset()
		code, err = cli.Run(os.Stdin, output, []string{
			"pgpst-cli",
			"db",
			"migrate",
			"--yes",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		// "Your schema is up to date" migration
		output.Reset()
		code, err = cli.Run(os.Stdin, output, []string{
			"pgpst-cli",
			"db",
			"migrate",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		// Create a new account
		input.Reset()
		input.WriteString(`{
	"main_address": "test123x",
	"password": "test123x",
	"subscription": "beta",
	"alt_email": "test123x@example.org",
	"status": "active"
}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		// Invalid JSON input in account creation
		input.Reset()
		input.WriteString(`{@@@@@}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Dry-create a new account using invalid manual inputs
		input.Reset()
		input.WriteString(`test123x
test123y
beta
test123y@example.org
inactive
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)
		input.Reset()

		input.WriteString(`test123y
test123y
betas
test123y@example.org
inactive
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(`test123y
test123y
beta
test123y@@@@
inactive
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(`test123y
test123y
beta
test123y@example.org
inactived
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		/*
		   		Convey("accs add --json and accs add should succeed", func() {
		   			jsonInput := strings.NewReader(`{
		   	"main_address": "test123x",
		   	"password": "test123x"
		   	"subscription": "beta",
		   	"alt_email": "test123x@example.org",
		   	"status": "active"
		   }`)
		   			jsonOutput := &bytes.Buffer{}
		   			code, err := cli.Run(jsonInput, jsonOutput, []string{
		   				"pgpst-cli",
		   				"accs",
		   				"add",
		   				"--json",
		   			})
		   			So(code, ShouldEqual, 0)
		   			So(err, ShouldBeNil)
		   			So(jsonOutput.String(), ShouldEqual, "Created a new account with ID")
		   		})*/
	})
}
