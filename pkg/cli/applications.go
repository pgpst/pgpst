package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/asaskevich/govalidator"
	"github.com/pgpst/pgpst/internal/github.com/codegangsta/cli"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/termtables"

	"github.com/pgpst/pgpst/pkg/models"
)

func applicationsAdd(c *cli.Context) {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return
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
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			writeError(err)
			return
		}
	} else {
		// Buffer stdin
		rd := bufio.NewReader(os.Stdin)
		var err error

		// Acquire from interactive input
		fmt.Print("Owner's ID: ")
		input.Owner, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return
		}
		input.Owner = strings.TrimSpace(input.Owner)

		fmt.Print("Application's name: ")
		input.Name, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return
		}
		input.Name = strings.TrimSpace(input.Name)

		fmt.Print("Homepage URL: ")
		input.Homepage, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return
		}
		input.Homepage = strings.TrimSpace(input.Homepage)

		fmt.Print("Description: ")
		input.Description, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return
		}
		input.Description = strings.TrimSpace(input.Description)

		fmt.Print("Callback URL: ")
		input.Callback, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return
		}
		input.Callback = strings.TrimSpace(input.Callback)
	}

	// Validate the input

	// Homepage URL should be a URL
	if !govalidator.IsURL(input.Homepage) {
		writeError(fmt.Errorf("%s is not a URL", input.Homepage))
		return
	}

	// Callback URL should be a URL
	if !govalidator.IsURL(input.Callback) {
		writeError(fmt.Errorf("%s is not a URL", input.Callback))
		return
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
		return
	}
	if !exists {
		writeError(fmt.Errorf("Account %s doesn't exist", input.Owner))
		return
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
		return
	}

	// Write a success message
	fmt.Printf("Created a new application with ID %s\n", application.ID)
}

func applicationsList(c *cli.Context) {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return
	}

	// Get applications from database
	cursor, err := r.Table("applications").Map(func(row r.Term) r.Term {
		return row.Merge(map[string]interface{}{
			"owners_address": r.Table("accounts").Get(row.Field("owner")).Field("main_address"),
		})
	}).Run(session)
	if err != nil {
		writeError(err)
		return
	}
	var applications []struct {
		models.Application
		OwnersAddress string `gorethink:"owners_address" json:"owners_address"`
	}
	if err := cursor.All(&applications); err != nil {
		writeError(err)
		return
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(os.Stdout).Encode(applications); err != nil {
			writeError(err)
			return
		}

		fmt.Print("\n")
		return
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
		return
	}
}
