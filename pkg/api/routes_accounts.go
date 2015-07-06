package api

import (
	"github.com/pgpst/pgpst/internal/github.com/asaskevich/govalidator"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

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
		// Normalize the username
		nu := utils.NormalizeUsername(input.Username)

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
		cursor, err := r.Table("addresses").GetAll(nu + "@pgp.st").Count().Eq(1).Do(func(left r.Term) map[string]interface{} {
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
	case "activate":
	}

	// Same as default in the switch
	c.JSON(422, &gin.H{
		"code":    0,
		"message": "Validation failed",
		"errors":  []string{"Invalid action"},
	})
	return
}
