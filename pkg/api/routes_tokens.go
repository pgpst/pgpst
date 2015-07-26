package api

import (
	"encoding/hex"
	"time"

	//"github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/asaskevich/govalidator"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
	"github.com/pgpst/pgpst/pkg/utils"
)

func (a *API) createToken(c *gin.Context) {
	// Decode the input
	var input struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		RedirectURI  string `json:"redirect_uri"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Address      string `json:"username"`
		Password     string `json:"password"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	// Switch the action
	switch input.GrantType {
	case "authorization_code":
		// Parameters:
		//  - code          - authorization code from the app
		//  - redirect_uri  - url used for the redirect, must be same
		//  - client_id     - id of the client app
		//  - client_secret - secret of the client app
	case "password":
		// Parameters:
		//  - username  - account's username
		//  - password  - sha256 of the account's password
		//  - client_id - id of the client app used for stats

		// Normaliuze the username
		na := utils.NormalizeAddress(input.Address)

		// Validate input
		errors := []string{}
		if !govalidator.IsEmail(na) {
			errors = append(errors, "Invalid address format.")
		}
		var dp []byte
		if len(input.Password) != 64 {
			errors = append(errors, "Invalid password length.")
		} else {
			var err error
			dp, err = hex.DecodeString(input.Password)
			if err != nil {
				errors = append(errors, "Invalid password format.")
			}
		}
		if len(errors) > 0 {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Validation failed.",
				"errors":  errors,
			})
			return
		}

		// Fetch the address from the database
		cursor, err := r.Table("addresses").Get(na).Do(func(address r.Term) map[string]interface{} {
			return map[string]interface{}{
				"address": address,
				"account": r.Table("accounts").Get(address.Field("owner")),
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
			Address *models.Address `gorethink:"address"`
			Account *models.Account `gorethink:"account"`
		}
		if err := cursor.One(&result); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Verify the password
		valid, update, err := result.Account.VerifyPassword(dp)
		if err != nil {
			c.JSON(401, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		if update {
			result.Account.DateModified = time.Now()
			if err := r.Table("accounts").Get(result.Account.ID).Update(result.Account).Exec(a.Rethink); err != nil {
				c.JSON(500, &gin.H{
					"code":    0,
					"message": err.Error(),
				})
				return
			}
		}
		if !valid {
			c.JSON(401, &gin.H{
				"code":    0,
				"message": "Invalid password",
			})
			return
		}
	case "client_credentials":
		// Parameters:
		//  - client_id     - id of the application
		//  - client_secret - secret of the application
	}

	// Same as default in the switch
	c.JSON(422, &gin.H{
		"code":    0,
		"message": "Validation failed",
		"errors":  []string{"Invalid action"},
	})
	return
}
