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

		// Write a response
		c.JSON(201, account)
		return
	case "activate":
		// Parameters:
		//  - address - expected address
		//  - token   - relevant token for address

		errors := []string{}

		if !govalidator.IsEmail(input.Address) {
			errors = append(errors, "Invalid address format")
		}
		if input.Token == "" {
			errors = append(errors, "Token is missing")
		}
		if len(errors) > 0 {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Validation failed",
				"errors":  errors,
			})
			return
		}

		// Normalise input.Address
		input_address := utils.RemoveDots(utils.NormalizeAddress(input.Address))

		// Check in the database whether these both exist
		cursor, err := r.Table("tokens").Get(input.Token).Do(func(left r.Term) r.Term {
			return r.Branch(
				left.Ne(nil).And(left.Field("type").Eq("activate")),
				map[string]interface{}{
					"token":   left,
					"account": r.Table("accounts").Get(left.Field("owner")),
				},
				map[string]interface{}{
					"token":   nil,
					"account": nil,
				},
			)
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
			Token   *models.Token   `gorethink:"token"`
			Account *models.Account `gorethink:"account"`
		}
		if err := cursor.One(&result); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		if result.Token == nil {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Activation failed",
				"errors":  []string{"Invalid token"},
			})
			return
		}

		if result.Account.MainAddress != input_address {
			errors = append(errors, "Address does not match token")
		} else if result.Account.Status != "inactive" {
			// Don't tell them an account is active is the address doesn't map
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
		if err := r.Table("tokens").Get(result.Token.ID).Delete().Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// and make the account active, welcome to pgp.st, new guy!
		if err := r.Table("accounts").Get(result.Account.ID).Update(map[string]interface{}{
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
			"id":      result.Account.ID,
			"message": "Activation successful",
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
