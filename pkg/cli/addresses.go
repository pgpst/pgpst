package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/codegangsta/cli"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/termtables"

	"github.com/pgpst/pgpst/pkg/models"
	"github.com/pgpst/pgpst/pkg/utils"
)

func addressesAdd(c *cli.Context) {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return
	}

	// Input struct
	var input struct {
		ID    string `json:"id"`
		Owner string `json:"owner"`
	}

	// Read JSON from stdin
	if c.Bool("json") {
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			writeError(err)
			return
		}
	} else {
		// Acquire from interactive input
		fmt.Print("Address: ")
		if _, err := fmt.Scanln(&input.ID); err != nil {
			writeError(err)
			return
		}

		fmt.Print("Owner ID: ")
		if _, err := fmt.Scanln(&input.Owner); err != nil {
			writeError(err)
			return
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
		return
	}
	defer cursor.Close()
	var taken bool
	if err := cursor.One(&taken); err != nil {
		writeError(err)
		return
	}
	if taken {
		writeError(fmt.Errorf("Address %s is already taken", input.ID))
		return
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
		return
	}
	if !exists {
		writeError(fmt.Errorf("Account %s doesn't exist", input.ID))
		return
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
		return
	}

	// Write a success message
	fmt.Printf("Created a new address - %s\n", address.ID)
}

func addressesList(c *cli.Context) {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return
	}

	// Get addresses from database
	cursor, err := r.Table("addresses").Map(func(row r.Term) r.Term {
		return row.Merge(map[string]interface{}{
			"main_address": r.Table("accounts").Get(row.Field("owner")).Field("main_address"),
		})
	}).Run(session)
	if err != nil {
		writeError(err)
		return
	}
	var addresses []struct {
		models.Address
		MainAddress string `gorethink:"main_address"`
	}
	if err := cursor.All(&addresses); err != nil {
		writeError(err)
		return
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(os.Stdout).Encode(addresses); err != nil {
			writeError(err)
			return
		}

		fmt.Print("\n")
		return
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
		return
	}
}
