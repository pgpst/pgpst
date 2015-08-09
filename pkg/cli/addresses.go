package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/cli"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/termtables"

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
			writeError(c, err)
			return 1
		}
	} else {
		// Buffer stdin
		rd := bufio.NewReader(c.App.Env["reader"].(io.Reader))
		var err error

		// Acquire from interactive input
		fmt.Fprintf(c.App.Writer, "Address: ")
		input.ID, err = rd.ReadString('\n')
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.ID = strings.TrimSpace(input.ID)

		fmt.Fprintf(c.App.Writer, "Owner ID: ")
		input.Owner, err = rd.ReadString('\n')
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.Owner = strings.TrimSpace(input.Owner)
	}

	// First of all, the address. Append domain if it has no such suffix.
	if strings.Index(input.ID, "@") == -1 {
		input.ID += "@" + c.GlobalString("default_domain")
	}

	// And format it
	styledID := utils.NormalizeAddress(input.ID)
	input.ID = utils.RemoveDots(styledID)

	// Then check if it's taken.
	cursor, err := r.Table("addresses").Get(input.ID).Ne(nil).Run(session)
	if err != nil {
		writeError(c, err)
		return 1
	}
	defer cursor.Close()
	var taken bool
	if err := cursor.One(&taken); err != nil {
		writeError(c, err)
		return 1
	}
	if taken {
		writeError(c, fmt.Errorf("Address %s is already taken", input.ID))
		return 1
	}

	// Check if account ID exists
	cursor, err = r.Table("accounts").Get(input.Owner).Ne(nil).Run(session)
	if err != nil {
		writeError(c, err)
	}
	defer cursor.Close()
	var exists bool
	if err := cursor.One(&exists); err != nil {
		writeError(c, err)
		return 1
	}
	if !exists {
		writeError(c, fmt.Errorf("Account %s doesn't exist", input.ID))
		return 1
	}

	// Insert the address into the database
	address := &models.Address{
		ID:           input.ID,
		StyledID:     styledID,
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        input.Owner,
	}

	if !c.GlobalBool("dry") {
		if err := r.Table("addresses").Insert(address).Exec(session); err != nil {
			writeError(c, err)
			return 1
		}
	}

	// Write a success message
	fmt.Fprintf(c.App.Writer, "Created a new address - %s (styled %s)\n", address.ID, address.StyledID)
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
		writeError(c, err)
		return 1
	}
	var addresses []struct {
		models.Address
		MainAddress string `gorethink:"main_address" json:"main_address"`
	}
	if err := cursor.All(&addresses); err != nil {
		writeError(c, err)
		return 1
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(c.App.Writer).Encode(addresses); err != nil {
			writeError(c, err)
			return 1
		}

		fmt.Fprintf(c.App.Writer, "\n")
	} else {
		table := termtables.CreateTable()
		table.AddHeaders("address", "styled_address", "main_address", "date_created")
		for _, address := range addresses {
			table.AddRow(
				address.ID,
				address.StyledID,
				address.MainAddress,
				address.DateCreated.Format(time.RubyDate),
			)
		}
		fmt.Fprintln(c.App.Writer, table.Render())
	}

	return 0
}
