package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pzduniak/termtables"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/cli"

	"github.com/pgpst/pgpst/pkg/models"
	"github.com/pgpst/pgpst/pkg/utils"
)

func addressesAdd(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Input struct
	var input struct {
		ID    string `json:"id"`
		Owner string `json:"owner"`
	}

	// Read JSON from stdin
	if c.Bool("json") {
		if err := json.NewDecoder(c.App.Env["reader"].(io.Reader)).Decode(&input); err != nil {
			writeError(err)
			return 1
		}
	} else {
		// Acquire from interactive input
		fmt.Print("Address: ")
		if _, err := fmt.Scanln(&input.ID); err != nil {
			writeError(err)
			return 1
		}

		fmt.Print("Owner ID: ")
		if _, err := fmt.Scanln(&input.Owner); err != nil {
			writeError(err)
			return 1
		}
	}

	// First of all, the address. Append domain if it has no such suffix.
	if strings.Index(input.ID, "@") == -1 {
		input.ID += "@" + c.String("default_domain")
	}

	// And format it
	input.ID = utils.NormalizeAddress(input.ID)

	// Then check if it's taken.
	cursor, err := r.Table("addresses").Get(input.ID).Ne(nil).Run(session)
	if err != nil {
		writeError(err)
		return 1
	}
	defer cursor.Close()
	var taken bool
	if err := cursor.One(&taken); err != nil {
		writeError(err)
		return 1
	}
	if taken {
		writeError(fmt.Errorf("Address %s is already taken", input.ID))
		return 1
	}

	// Check if account ID exists
	cursor, err = r.Table("accounts").Get(input.Owner).Ne(nil).Run(session)
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
		writeError(fmt.Errorf("Account %s doesn't exist", input.ID))
		return 1
	}

	// Insert the address into the database
	address := &models.Address{
		ID:           input.ID,
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        input.Owner,
	}
	if err := r.Table("addresses").Insert(address).Exec(session); err != nil {
		writeError(err)
		return 1
	}

	// Write a success message
	fmt.Printf("Created a new address - %s\n", address.ID)
	return 0
}

func addressesList(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Get addresses from database
	cursor, err := r.Table("addresses").Map(func(row r.Term) r.Term {
		return row.Merge(map[string]interface{}{
			"main_address": r.Table("accounts").Get(row.Field("owner")).Field("main_address"),
		})
	}).Run(session)
	if err != nil {
		writeError(err)
		return 1
	}
	var addresses []struct {
		models.Address
		MainAddress string `gorethink:"main_address" json:"main_address"`
	}
	if err := cursor.All(&addresses); err != nil {
		writeError(err)
		return 1
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(c.App.Writer).Encode(addresses); err != nil {
			writeError(err)
			return 1
		}

		fmt.Print("\n")
	} else {
		table := termtables.CreateTable()
		table.AddHeaders("address", "main_addresss", "date_created")
		for _, address := range addresses {
			table.AddRow(
				address.ID,
				address.MainAddress,
				address.DateCreated.Format(time.RubyDate),
			)
		}
		fmt.Println(table.Render())
	}

	return 0
}
