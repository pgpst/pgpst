package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/termtables"
	"github.com/pzduniak/cli"

	"github.com/pgpst/pgpst/pkg/models"
)

func tokensAdd(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Input struct
	var input struct {
		Owner      string    `json:"owner"`
		ExpiryDate time.Time `json:"expiry_date"`
		Type       string    `json:"type"`
		Scope      []string  `json:"scope"`
		ClientID   string    `json:"client_id"`
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

		fmt.Print("Type [auth/activate]: ")
		input.Type, err = rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return 1
		}
		input.Type = strings.TrimSpace(input.Type)

		fmt.Print("Expiry date [2006-01-02T15:04:05Z07:00/empty]: ")
		expiryDate, err := rd.ReadString('\n')
		if err != nil {
			writeError(err)
			return 1
		}
		expiryDate = strings.TrimSpace(expiryDate)
		if expiryDate != "" {
			input.ExpiryDate, err = time.Parse(time.RFC3339, expiryDate)
			if err != nil {
				writeError(err)
				return 1
			}
		}

		if input.Type == "auth" {
			fmt.Print("Client ID: ")
			input.ClientID, err = rd.ReadString('\n')
			if err != nil {
				writeError(err)
				return 1
			}
			input.ClientID = strings.TrimSpace(input.ClientID)

			fmt.Print("Scope (seperated by commas): ")
			scope, err := rd.ReadString('\n')
			if err != nil {
				writeError(err)
				return 1
			}
			scope = strings.TrimSpace(scope)
			input.Scope = strings.Split(scope, ",")
		}
	}

	// Validate the input

	// Type has to be either auth or activate
	if input.Type != "auth" && input.Type != "activate" {
		writeError(fmt.Errorf("Token type must be either auth or activate. Got %s.", input.Type))
		return 1
	}

	// Scopes must exist
	if input.Scope != nil && len(input.Scope) > 0 {
		for _, scope := range input.Scope {
			if _, ok := models.Scopes[scope]; !ok {
				writeError(fmt.Errorf("Scope %s doesn't exist", scope))
				return 1
			}
		}
	}

	// Owner must exist
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

	// Application must exist
	if input.ClientID != "" {
		cursor, err = r.Table("applications").Get(input.ClientID).Ne(nil).Run(session)
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
			writeError(fmt.Errorf("Application %s doesn't exist", input.ClientID))
			return 1
		}
	}

	// Insert into database
	token := &models.Token{
		ID:           uniuri.NewLen(uniuri.UUIDLen),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        input.Owner,
		ExpiryDate:   input.ExpiryDate,
		Type:         input.Type,
		Scope:        input.Scope,
		ClientID:     input.ClientID,
	}
	if err := r.Table("tokens").Insert(token).Exec(session); err != nil {
		writeError(err)
		return 1
	}

	// Write a success message
	fmt.Printf("Created a new %s token with ID %s\n", token.Type, token.ID)
	return 0
}

func tokensList(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Get tokens from database
	cursor, err := r.Table("tokens").Map(func(row r.Term) r.Term {
		return r.Branch(
			row.HasFields("client_id"),
			row.Merge(map[string]interface{}{
				"owners_address": r.Table("accounts").Get(row.Field("owner")).Field("main_address"),
				"client_name":    r.Table("applications").Get(row.Field("client_id")).Field("name"),
			}),
			row.Merge(map[string]interface{}{
				"owners_address": r.Table("accounts").Get(row.Field("owner")).Field("main_address"),
			}),
		)
	}).Run(session)
	if err != nil {
		writeError(err)
		return 1
	}
	var tokens []struct {
		models.Token
		OwnersAddress string `gorethink:"owners_address" json:"owners_address"`
		ClientName    string `gorethink:"client_name" json:"client_name,omitempty"`
	}
	if err := cursor.All(&tokens); err != nil {
		writeError(err)
		return 1
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(c.App.Writer).Encode(tokens); err != nil {
			writeError(err)
			return 1
		}

		fmt.Print("\n")
	} else {
		table := termtables.CreateTable()
		table.AddHeaders("id", "type", "owner", "client_name", "expired", "date_created")
		for _, token := range tokens {
			table.AddRow(
				token.ID,
				token.Type,
				token.OwnersAddress,
				token.ClientName,
				!token.ExpiryDate.IsZero() && token.ExpiryDate.Before(time.Now()),
				token.DateCreated.Format(time.RubyDate),
			)
		}
		fmt.Println(table.Render())
	}

	return 0
}
