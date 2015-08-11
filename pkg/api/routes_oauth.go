package api

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/asaskevich/govalidator"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
	"github.com/pgpst/pgpst/pkg/utils"
)

func (a *API) oauthToken(c *gin.Context) {
	// Decode the input
	var input struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Address      string `json:"username"`
		Password     string `json:"password"`
		ExpiryTime   int64  `json:"expiry_time"`
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
		//  - client_id     - id of the client app
		//  - client_secret - secret of the client app

		// Fetch the application from database
		cursor, err := r.Table("applications").Get(input.ClientID).Default(map[string]interface{}{}).Run(a.Rethink)
		if err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		defer cursor.Close()
		var application *models.Application
		if err := cursor.One(&application); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		if application.ID == "" {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "No such client ID.",
			})
			return
		}
		if application.Secret != input.ClientSecret {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Invalid client secret.",
			})
			return
		}

		// Fetch the code from the database
		cursor, err = r.Table("tokens").Get(input.Code).Default(map[string]interface{}{}).Run(a.Rethink)
		if err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		defer cursor.Close()
		var codeToken *models.Token
		if err := cursor.One(&codeToken); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Ensure token type and matching client id
		if codeToken.ID == "" || codeToken.Type != "code" || codeToken.ClientID != input.ClientID {
			c.JSON(422, &gin.H{
				"code":    0,
				"message": "Invalid code",
			})
			return
		}

		// Create a new authentication code
		token := &models.Token{
			ID:           uniuri.NewLen(uniuri.UUIDLen),
			DateCreated:  time.Now(),
			DateModified: time.Now(),
			Owner:        codeToken.Owner,
			ExpiryDate:   codeToken.ExpiryDate,
			Type:         "auth",
			Scope:        codeToken.Scope,
			ClientID:     input.ClientID,
		}

		// Remove code token
		if err := r.Table("tokens").Get(codeToken.ID).Delete().Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Insert it into database
		if err := r.Table("tokens").Insert(token).Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Write the token into the response
		c.JSON(201, token)
		return
	case "password":
		// Parameters:
		//  - username    - account's username
		//  - password    - sha256 of the account's password
		//  - client_id   - id of the client app used for stats
		//  - expiry_time - seconds until token expires

		// If there's no domain, append default domain
		if strings.Index(input.Address, "@") == -1 {
			input.Address += "@" + a.Options.DefaultDomain
		}

		// Normalize the username
		na := utils.RemoveDots(utils.NormalizeAddress(input.Address))

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
		if input.ExpiryTime == 0 {
			input.ExpiryTime = 86400 // 24 hours
		} else if input.ExpiryTime < 0 {
			errors = append(errors, "Invalid expiry time.")
		}
		if input.ClientID == "" {
			errors = append(errors, "Missing client ID.")
		} else {
			cursor, err := r.Table("applications").Get(input.ClientID).Ne(nil).Run(a.Rethink)
			if err != nil {
				c.JSON(500, &gin.H{
					"code":    0,
					"message": err.Error(),
				})
				return
			}
			defer cursor.Close()
			var appExists bool
			if err := cursor.One(&appExists); err != nil {
				c.JSON(500, &gin.H{
					"code":    0,
					"message": err.Error(),
				})
				return
			}
			if !appExists {
				errors = append(errors, "Invalid client ID.")
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

		// Create a new token
		token := &models.Token{
			ID:           uniuri.NewLen(uniuri.UUIDLen),
			DateCreated:  time.Now(),
			DateModified: time.Now(),
			Owner:        result.Account.ID,
			ExpiryDate:   time.Now().Add(time.Duration(input.ExpiryTime) * time.Second),
			Type:         "auth",
			Scope:        []string{"password_grant"},
			ClientID:     input.ClientID,
		}

		if result.Account.Subscription == "admin" {
			token.Scope = append(token.Scope, "admin")
		}

		// Insert it into database
		if err := r.Table("tokens").Insert(token).Exec(a.Rethink); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Write the token into the response
		c.JSON(201, token)
		return
	case "client_credentials":
		// Parameters:
		//  - client_id     - id of the application
		//  - client_secret - secret of the application

		c.JSON(501, &gin.H{
			"code":    0,
			"message": "Client credentials flow is not implemented.",
		})
	}

	// Same as default in the switch
	c.JSON(422, &gin.H{
		"code":    0,
		"message": "Validation failed",
		"errors":  []string{"Invalid action"},
	})
	return
}
