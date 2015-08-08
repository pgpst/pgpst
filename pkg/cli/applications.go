package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/asaskevich/govalidator"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pzduniak/termtables"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/cli"

	"github.com/pgpst/pgpst/pkg/models"
)

func applicationsAdd(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Input struct
	var input struct {
		Owner       string `json:"owner"`
		Callback    string `json:"callback"`
		Homepage    string `json:"homepage"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	// Read JSON from stdin
	if c.Bool("json") {
		if err := json.NewDecoder(c.App.Env["reader"].(io.Reader)).Decode(&input); err != nil {
			writeError(err)
			return 1
		}
	} else {
		// Buffer stdin
		rd := bufio.NewReader(c.App.Env["reader"].(io.Reader))
		var err error

		// Acquire from interactive input
		fmt.Print("Owner's ID: ")
		input.Owner, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return 1
		}
		input.Owner = strings.TrimSpace(input.Owner)

		fmt.Print("Application's name: ")
		input.Name, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return 1
		}
		input.Name = strings.TrimSpace(input.Name)

		fmt.Print("Homepage URL: ")
		input.Homepage, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return 1
		}
		input.Homepage = strings.TrimSpace(input.Homepage)

		fmt.Print("Description: ")
		input.Description, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return 1
		}
		input.Description = strings.TrimSpace(input.Description)

		fmt.Print("Callback URL: ")
		input.Callback, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return 1
		}
		input.Callback = strings.TrimSpace(input.Callback)
	}

	// Validate the input

	// Homepage URL should be a URL
	if !govalidator.IsURL(input.Homepage) {
		writeError(fmt.Errorf("%s is not a URL", input.Homepage))
		return 1
	}

	// Callback URL should be a URL
	if !govalidator.IsURL(input.Callback) {
		writeError(fmt.Errorf("%s is not a URL", input.Callback))
		return 1
	}

	// Check if account ID exists
	cursor, err := r.Table("accounts").Get(input.Owner).Ne(nil).Run(session)
	if err != nil {
		writeError(err)
	}
	defer cursor.Close()
	var exists bool
	if err := cursor.One(&exists); err != nil {
		writeError(err)
		return 1
	}
	if !exists {
		writeError(fmt.Errorf("Account %s doesn't exist", input.Owner))
		return 1
	}

	// Insert into database
	application := &models.Application{
		ID:           uniuri.NewLen(uniuri.UUIDLen),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        input.Owner,
		Secret:       uniuri.NewLen(32),
		Callback:     input.Callback,
		Homepage:     input.Homepage,
		Name:         input.Name,
		Description:  input.Description,
	}
	if err := r.Table("applications").Insert(application).Exec(session); err != nil {
		writeError(err)
		return 1
	}

	// Write a success message
	fmt.Printf("Created a new application with ID %s\n", application.ID)
	return 0
}

func applicationsList(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Get applications from database
	cursor, err := r.Table("applications").Map(func(row r.Term) r.Term {
		return row.Merge(map[string]interface{}{
			"owners_address": r.Table("accounts").Get(row.Field("owner")).Field("main_address"),
		})
	}).Run(session)
	if err != nil {
		writeError(err)
		return 1
	}
	var applications []struct {
		models.Application
		OwnersAddress string `gorethink:"owners_address" json:"owners_address"`
	}
	if err := cursor.All(&applications); err != nil {
		writeError(err)
		return 1
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(c.App.Writer).Encode(applications); err != nil {
			writeError(err)
			return 1
		}

		fmt.Print("\n")
	} else {
		table := termtables.CreateTable()
		table.AddHeaders("id", "name", "owner", "homepage", "date_created")
		for _, application := range applications {
			table.AddRow(
				application.ID,
				application.Name,
				application.OwnersAddress,
				application.Homepage,
				application.DateCreated.Format(time.RubyDate),
			)
		}
		fmt.Println(table.Render())
	}

	return 0
}
