package cli

import (
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/pgpst/pgpst/pkg/utils"
)

func accountsAdd(c *cli.Context) {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return
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
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			writeError(err)
			return
		}
	} else {
		// Acquire from interactive input
		fmt.Print("Main address: ")
		if _, err := fmt.Scanln(&input.MainAddress); err != nil {
			writeError(err)
			return
		}

		password, err := utils.AskPassword("Password: ")
		if err != nil {
			writeError(err)
			return
		}
		input.Password = password

		fmt.Print("Subscription [beta/admin]: ")
		if _, err := fmt.Scanln(&input.Subscription); err != nil {
			writeError(err)
			return
		}

		fmt.Print("Alternative address: ")
		if _, err := fmt.Scanln(&input.AltEmail); err != nil {
			writeError(err)
			return
		}

		fmt.Print("Status [inactive/active/suspended]: ")
		if _, err := fmt.Scanln(&input.Status); err != nil {
			writeError(err)
			return
		}
	}

	// Analyze the input

	// First of all, the address. Append domain if it has no such suffix.
	if strings.Index(input.MainAddress, "@") == -1 {
		input.MainAddress += "@" + c.String("default_domain")
	}

	// And format it
	input.MainAddress = utils.NormalizeAddress(input.MainAddress)

	// Then check if it's taken.
	cursor, err := r.Table("addresses").Get(input.MainAddress).Ne(nil).Run(session)
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
		writeError(fmt.Errorf("Address %s is already taken", input.MainAddress))
		return
	}

	// If the password isn't 64 characters long, then hash it.
	if len(input.Password) != 64 {
		hash := sha256.Sum256([]byte(input.Password))
		input.Password = hex.EncodeToString(hash[:])
	}

	// Subscription has to be beta or admin
	if input.Subscription != "beta" && input.Subscription != "admin" {
		writeError(fmt.Errorf("Subscription has to be either beta or admin. Got %s.", input.Subscription))
		return
	}

	// AltEmail must be an email
	if !govalidator.IsEmail(input.AltEmail) {
		writeError(fmt.Errorf("Email %s has an incorrect format", input.AltEmail))
		return
	}

	// Status has to be inactive/active/suspended
	if input.Status != "inactive" && input.Status != "active" && input.Status != "suspended" {
		writeError(fmt.Errorf("Status has to be either inactive, active or suspended. Got %s.", input.Status))
		return
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
		writeError(err)
		return
	}

	address := &models.Address{
		ID:           input.MainAddress,
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        account.ID,
	}

	// Insert them into database
	if err := r.Table("addresses").Insert(address).Exec(session); err != nil {
		writeError(err)
		return
	}
	if err := r.Table("accounts").Insert(account).Exec(session); err != nil {
		writeError(err)
		return
	}

	// Write a success message
	fmt.Printf("Created a new account with ID %s\n", account.ID)
}

func accountsList(c *cli.Context) {
	// Connect to RethinkDB
	_, session, connected := connectToRethinkDB(c)
	if !connected {
		return
	}

	// Get accounts without passwords from database
	cursor, err := r.Table("accounts").Map(func(row r.Term) r.Term {
		return row.Without("password").Merge(map[string]interface{}{
			"addresses": r.Table("addresses").GetAllByIndex("owner", row.Field("id")).CoerceTo("array"),
		})
	}).Run(session)
	if err != nil {
		writeError(err)
		return
	}
	var accounts []struct {
		models.Account
		Addresses []*models.Address `gorethink:"addresses"`
	}
	if err := cursor.All(&accounts); err != nil {
		writeError(err)
		return
	}

	// Write the output
	if c.Bool("json") {
		if err := json.NewEncoder(os.Stdout).Encode(accounts); err != nil {
			writeError(err)
			return
		}

		fmt.Print("\n")
		return
	} else {
		table := termtables.CreateTable()
		table.AddHeaders("id", "addresses", "subscription", "status", "date_created")
		for _, account := range accounts {
			emails := []string{}

			for _, address := range account.Addresses {
				if address.ID == account.MainAddress {
					address.ID = "*" + address.ID
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
		fmt.Println(table.Render())
		return
	}
}
