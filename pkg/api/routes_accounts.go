package api

import (
	"time"

	"github.com/pgpst/pgpst/internal/github.com/asaskevich/govalidator"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
	"github.com/pgpst/pgpst/pkg/utils"
)

func (a *API) createAccount(c *gin.Context) {
	// Decode the input
	var input struct {
		Action   string `json:"action"`
		Username string `json:"username"`
		AltEmail string `json:"alt_email"`
		Token    string `json:"token"`
		Password string `json:"password"`
		Address  string `json:"address"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	// Switch the action
	switch input.Action {
	case "reserve":
		// Parameters:
		//  - username  - desired username
		//  - alt_email - desired email

		// Normalize the username
		nu := utils.RemoveDots(utils.NormalizeUsername(input.Username))

		// Validate input:
		// - len(username) >= 3 && len(username) <= 32
		// - email.match(alt_email)
		errors := []string{}
		if len(nu) < 3 {
			errors = append(errors, "Username too short. It must be 3-32 characters long.")
		}
		if len(nu) > 32 {
			errors = append(errors, "Username too long. It must be 3-32 characters long.")
		}
		if !govalidator.IsEmail(input.AltEmail) {
			errors = append(errors, "Invalid alternative e-mail format.")
		}
		if len(errors) > 0 {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Validation failed.",
				"errors":  errors,
			})
			return
		}

		// Check in the database whether you can register such account
		cursor, err := r.Table("addresses").Get(nu + "@pgp.st").Ne(nil).Do(func(left r.Term) map[string]interface{} {
			return map[string]interface{}{
				"username":  left,
				"alt_email": r.Table("accounts").GetAllByIndex("alt_email", input.AltEmail).Count().Eq(1),
			}
		}).Run(a.Rethink)
		if err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		defer cursor.Close()
		var result struct {
			Username bool `gorethink:"username"`
			AltEmail bool `gorethink:"alt_email"`
		}
		if err := cursor.One(&result); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		if result.Username || result.AltEmail {
			errors := []string{}
			if result.Username {
				errors = append(errors, "This username is taken.")
			}
			if result.AltEmail {
				errors = append(errors, "This email address is used.")
			}
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Naming conflict",
				"errors":  errors,
			})
			return
		}

		// Create an account and an address
		address := &models.Address{
			ID:          nu + "@pgp.st",
			DateCreated: time.Now(),
			Owner:       "", // we set it later
		}
		account := &models.Account{
			ID:           uniuri.NewLen(uniuri.UUIDLen),
			DateCreated:  time.Now(),
			MainAddress:  address.ID,
			Subscription: "beta",
			AltEmail:     input.AltEmail,
			Status:       "inactive",
		}
		address.Owner = account.ID

		// Insert them into the database
		if err := r.Table("addresses").Insert(address).Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		if err := r.Table("accounts").Insert(account).Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		if err := r.Table("accounts").Insert(account).Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Write a response
		c.JSON(201, account)
		return
	case "activate":
		// Parameters:
		//  - address - expected address
		//  - token   - relevant token for address

		errors := []string{}

		// Normalise input.Address
		input_address := utils.NormalizeAddress(input.Address)

		if !govalidator.IsEmail(input_address) {
			errors = append(errors, "Invalid address format")
		}
		if input.Token == "" {
			errors = append(errors, "Invalid token - none given")
		}
		if len(errors) > 0 {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Validation failed",
				"errors":  errors,
			})
			return
		}

		// Check in the database whether these both exist
		cursor, err := r.Table("tokens").Get(input.Token).Ne(nil).Do(func(left r.Term) map[string]interface{} {
			owner := r.Table("tokens").Get(input.Token).Field("owner")

			return map[string]interface{}{
				"id": left,

				// If token exists, get the owner
				"owner": r.Branch(
					left.Eq(true),
					owner,
					"",
				),

				// If token exists, check if type="activate"
				"type": r.Branch(
					left.Eq(true),
					r.Table("tokens").Get(input.Token).Field("type").Eq("activate"),
					true, // return true even if token doesn't exist so that it doesnt errorspam
				),

				// If token exists, check if owner=input_address
				"address": r.Branch(
					left.Eq(true),
					r.Table("accounts").Get(owner).Field("main_address").Eq(input_address),
					true, // return true even if token doesn't exist so no errorspam
				),

				// Also check if owner is inactive
				"inactive": r.Branch(
					left.Eq(true),
					r.Table("accounts").Get(owner).Field("status").Eq("inactive"),
					true, // you get the point
				),
			}
		}).Run(a.Rethink)
		if err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		defer cursor.Close()
		var result struct {
			ID       bool   `gorethink:"id"`
			Owner    string `gorethink:"owner"`
			Type     bool   `gorethink:"type"`
			Address  bool   `gorethink:"address"`
			Inactive bool   `gorethink:"inactive"`
		}
		if err := cursor.One(&result); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		if !result.ID {
			errors = append(errors, "Invalid token - token doesn't exist")
		}
		if !result.Type {
			errors = append(errors, "Invalid token - wrong token type")
		}
		if !result.Address {
			errors = append(errors, "Address doesn't map to token")
		}
		if !result.Inactive {
			errors = append(errors, "Account already active")
		}

		if len(errors) > 0 {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Activation failed",
				"errors":  errors,
			})
			return
		}

		// Everything seems okay, let's go ahead and delete the token
		if err := r.Table("tokens").Get(input.Token).Delete().Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		// and make the account active, welcome to pgp.st, new guy!
		if err := r.Table("accounts").Get(result.Owner).Update(map[string]interface{}{
			"status": "active",
		}).Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Temporary response...
		c.JSON(201, &gin.H{
			"id":      result.Owner,
			"message": "Activation success",
		})
		return
	}

	// Same as default in the switch
	c.JSON(422, &gin.H{
		"code":    0,
		"message": "Validation failed",
		"errors":  []string{"Invalid action"},
	})
	return
}
