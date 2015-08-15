package api

import (
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) createToken(c *gin.Context) {
	// Get token and account info from the context
	var (
		account      = c.MustGet("account").(*models.Account)
		currentToken = c.MustGet("token").(*models.Token)
	)

	if !models.InScope(currentToken.Scope, []string{"tokens:oauth"}) {
		c.JSON(403, &gin.H{
			"code":  0,
			"error": "Your token has insufficient scope",
		})
		return
	}

	// Decode the input
	var input struct {
		Type       string   `json:"type"`
		ClientID   string   `json:"client_id"`
		Scope      []string `json:"scope"`
		ExpiryTime int64    `json:"expiry_time"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	// Run the validation
	errors := []string{}

	// Type has to be code
	if input.Type != "code" && input.Type != "auth" {
		errors = append(errors, "Only \"code\" token can be created using this endpoint.")
	}

	// Scope must contain proper scopes
	sm := map[string]struct{}{}
	for _, scope := range input.Scope {
		if _, ok := models.Scopes[scope]; !ok {
			errors = append(errors, "Scope \""+scope+"\" does not exist.")
		} else {
			sm[scope] = struct{}{}
		}
	}
	if _, ok := sm["password_grant"]; ok {
		errors = append(errors, "You can not request the password grant scope.")
	}
	if _, ok := sm["admin"]; ok && account.Subscription != "admin" {
		errors = append(errors, "You can not request the admin scope.")
	}

	// Expiry time must be valid
	if input.ExpiryTime == 0 {
		input.ExpiryTime = 86400
	} else if input.ExpiryTime < 0 {
		errors = append(errors, "Invalid expiry time.")
	}

	// Client ID has to be an application ID
	var application *models.Application
	if input.ClientID == "" {
		errors = append(errors, "Client ID is missing.")
	} else {
		cursor, err := r.Table("applications").Get(input.ClientID).Default(map[string]interface{}{}).Run(a.Rethink)
		if err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		defer cursor.Close()
		if err := cursor.One(&application); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		if application.ID == "" {
			errors = append(errors, "There is no such application.")
		}
	}

	// Abort the request if there are errors
	if len(errors) > 0 {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": "Validation failed.",
			"errors":  errors,
		})
		return
	}

	// Create a new token
	token := &models.Token{
		ID:           uniuri.NewLen(uniuri.UUIDLen),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        account.ID,
		ExpiryDate:   time.Now().Add(time.Duration(input.ExpiryTime) * time.Second),
		Type:         input.Type,
		Scope:        input.Scope,
		ClientID:     input.ClientID,
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
}
