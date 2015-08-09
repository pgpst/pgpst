package cli

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/asaskevich/govalidator"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/cli"
	"github.com/pgpst/pgpst/internal/github.com/pzduniak/termtables"

	"github.com/pgpst/pgpst/pkg/models"
	"github.com/pgpst/pgpst/pkg/utils"
)

func accountsAdd(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Input struct
	var input struct {
		MainAddress  string `json:"main_address"`
		Password     string `json:"password"`
		Subscription string `json:"subscription"`
		AltEmail     string `json:"alt_email"`
		Status       string `json:"status"`
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
		fmt.Fprint(c.App.Writer, "Main address: ")
		input.MainAddress, err = rd.ReadString('\n')
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.MainAddress = strings.TrimSpace(input.MainAddress)

		input.Password, err = rd.ReadString('\n')
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.Password = strings.TrimSpace(input.Password)

		/*password, err := speakeasy.FAsk(rd, "Password: ")
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.Password = password*/

		fmt.Fprint(c.App.Writer, "Subscription [beta/admin]: ")
		input.Subscription, err = rd.ReadString('\n')
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.Subscription = strings.TrimSpace(input.Subscription)

		fmt.Fprint(c.App.Writer, "Alternative address: ")
		input.AltEmail, err = rd.ReadString('\n')
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.AltEmail = strings.TrimSpace(input.AltEmail)

		fmt.Fprint(c.App.Writer, "Status [inactive/active/suspended]: ")
		input.Status, err = rd.ReadString('\n')
		if err != nil {
			writeError(c, err)
			return 1
		}
		input.Status = strings.TrimSpace(input.Status)
	}

	// Analyze the input

	// First of all, the address. Append domain if it has no such suffix.
	if strings.Index(input.MainAddress, "@") == -1 {
		input.MainAddress += "@" + c.GlobalString("default_domain")
	}

	// And format it
	styledID := utils.NormalizeAddress(input.MainAddress)
	input.MainAddress = utils.RemoveDots(styledID)

	// Then check if it's taken.
	cursor, err := r.Table("addresses").Get(input.MainAddress).Ne(nil).Run(session)
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
		writeError(c, fmt.Errorf("Address %s is already taken", input.MainAddress))
		return 1
	}

	// If the password isn't 64 characters long, then hash it.
	if len(input.Password) != 64 {
		hash := sha256.Sum256([]byte(input.Password))
		input.Password = hex.EncodeToString(hash[:])
	}

	// Subscription has to be beta or admin
	if input.Subscription != "beta" && input.Subscription != "admin" {
		writeError(c, fmt.Errorf("Subscription has to be either beta or admin. Got %s.", input.Subscription))
		return 1
	}

	// AltEmail must be an email
	if !govalidator.IsEmail(input.AltEmail) {
		writeError(c, fmt.Errorf("Email %s has an incorrect format", input.AltEmail))
		return 1
	}

	// Status has to be inactive/active/suspended
	if input.Status != "inactive" && input.Status != "active" && input.Status != "suspended" {
		writeError(c, fmt.Errorf("Status has to be either inactive, active or suspended. Got %s.", input.Status))
		return 1
	}

	// Prepare structs to insert
	account := &models.Account{
		ID:           uniuri.NewLen(uniuri.UUIDLen),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		MainAddress:  input.MainAddress,
		Subscription: input.Subscription,
		AltEmail:     input.AltEmail,
		Status:       input.Status,
	}
	if err := account.SetPassword([]byte(input.Password)); err != nil {
		writeError(c, err)
		return 1
	}

	address := &models.Address{
		ID:           input.MainAddress,
		StyledID:     styledID,
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        account.ID,
	}

	// Insert them into database
	if !c.Bool("dry") {
		if err := r.Table("addresses").Insert(address).Exec(session); err != nil {
			writeError(c, err)
			return 1
		}
		if err := r.Table("accounts").Insert(account).Exec(session); err != nil {
			writeError(c, err)
			return 1
		}
	}

	// Write a success message
	fmt.Fprintf(c.App.Writer, "Created a new account with ID %s\n", account.ID)
	return 0
}

func accountsList(c *cli.Context) int {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return 1
	}

	// Get accounts without passwords from database
	cursor, err := r.Table("accounts").Map(func(row r.Term) r.Term {
		return row.Without("password").Merge(map[string]interface{}{
			"addresses": r.Table("addresses").GetAllByIndex("owner", row.Field("id")).CoerceTo("array"),
		})
	}).Run(session)
	if err != nil {
		writeError(c, err)
		return 1
	}
	var accounts []struct {
		models.Account
		Addresses []*models.Address `gorethink:"addresses" json:"addresses`
	}
	if err := cursor.All(&accounts); err != nil {
		writeError(c, err)
		return 1
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(c.App.Writer).Encode(accounts); err != nil {
			writeError(c, err)
			return 1
		}

		fmt.Fprint(c.App.Writer, "\n")
	} else {
		table := termtables.CreateTable()
		table.AddHeaders("id", "addresses", "subscription", "status", "date_created")
		for _, account := range accounts {
			emails := []string{}

			for _, address := range account.Addresses {
				if address.ID == account.MainAddress {
					address.ID = fmt.Sprintf("* %s (styled: %s)", address.ID, address.StyledID)
					emails = append([]string{address.ID}, emails...)
				} else {
					emails = append(emails, address.ID)
				}
			}

			table.AddRow(
				account.ID,
				strings.Join(emails, ", "),
				account.Subscription,
				account.Status,
				account.DateCreated.Format(time.RubyDate),
			)
		}
		fmt.Fprintln(c.App.Writer, table.Render())
	}

	return 0
}
