package cli_test

import (
	"bytes"
	"os"
	"regexp"
	"testing"
	"time"

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

		// Match the ID
		expr := regexp.MustCompile(`Created a new account with ID (.*)\n`)
		accountID := expr.FindStringSubmatch(output.String())[1]
		So(accountID, ShouldNotBeEmpty)

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

		// Create a new application
		input.Reset()
		input.WriteString(`{
	"owner": "` + accountID + `",
	"callback": "https://example.org/callback",
	"homepage": "https://example.org",
	"name": "Example application",
	"description": "An example application created using a test"
}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"apps",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		expr = regexp.MustCompile(`Created a new application with ID (.*)\n`)
		applicationID := expr.FindStringSubmatch(output.String())[1]
		So(applicationID, ShouldNotBeEmpty)

		// Invalid JSON input in application creation
		input.Reset()
		input.WriteString(`{@@@@@}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"apps",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Dry-create a new application using invalid manual inputs
		input.Reset()
		input.WriteString(`ownerid
appname
homepageurl
description
callback
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"apps",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(accountID + `
appname
homepageurl::
description
callback
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"apps",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(accountID + `
appname
http://example.org
description
callback::
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"apps",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Create a new token
		input.Reset()
		input.WriteString(`{
	"owner": "` + accountID + `",
	"expiry_date": "` + time.Now().Add(time.Hour*24).Format(time.RFC3339) + `",
	"type": "auth",
	"scope": ["applications", "resources", "tokens"],
	"client_id": "` + applicationID + `"
}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		expr = regexp.MustCompile(`Created a new auth token with ID (.*)\n`)
		tokenID := expr.FindStringSubmatch(output.String())[1]
		So(applicationID, ShouldNotBeEmpty)

		// Invalid JSON input in token creation
		input.Reset()
		input.WriteString(`{@@@@@}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Dry-create a new token using invalid manual inputs
		input.Reset()
		input.WriteString(`ownerid
authed
notapropertimestring
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(`ownerid
authed
` + time.Now().Add(time.Hour*24).Format(time.RFC3339) + `
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(`ownerid
auth
` + time.Now().Add(time.Hour*24).Format(time.RFC3339) + `
clientid
notinscopes
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(`ownerid
authed
` + time.Now().Add(time.Hour*24).Format(time.RFC3339) + `
clientid
notinscopes
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(`ownerid
auth
` + time.Now().Add(time.Hour*24).Format(time.RFC3339) + `
clientid
account
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(accountID + `
auth
` + time.Now().Add(time.Hour*24).Format(time.RFC3339) + `
clientid
account
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Create a new address
		input.Reset()
		input.WriteString(`{
	"id": "test123123",
	"owner": "` + accountID + `"
}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"addrs",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)

		// Invalid input address creation
		input.Reset()
		input.WriteString(`{@@@@@}`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"addrs",
			"add",
			"--json",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Dry-create a new address using invalid manual inputs
		input.Reset()
		input.WriteString(`test123x
ownerid
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"addrs",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		input.Reset()
		input.WriteString(`test123123123
ownerid
`)
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"addrs",
			"add",
			"--dry",
		})
		So(code, ShouldEqual, 1)
		So(err, ShouldBeNil)

		// Check existence in list commands
		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"list",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, "\""+accountID+"\"")

		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"accs",
			"list",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, accountID)

		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"addrs",
			"list",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, "\"test123x@pgp.st\"")
		So(output.String(), ShouldContainSubstring, "\"test123123@pgp.st\"")

		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"addrs",
			"list",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, "test123x@pgp.st")
		So(output.String(), ShouldContainSubstring, "test123123@pgp.st")

		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"apps",
			"list",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, "\""+applicationID+"\"")

		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"apps",
			"list",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, applicationID)

		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"list",
			"--json",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, "\""+tokenID+"\"")

		output.Reset()
		code, err = cli.Run(input, output, []string{
			"pgpst-cli",
			"toks",
			"list",
		})
		So(code, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(output.String(), ShouldContainSubstring, tokenID)

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
