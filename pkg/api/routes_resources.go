package api

import (
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) createResource(c *gin.Context) {
	// Get token and account info from the context
	var (
		account = c.MustGet("account").(*models.Account)
		token   = c.MustGet("token").(*models.Token)
	)

	// Check the scope
	if !models.InScope(token.Scope, []string{"resources:create"}) {
		c.JSON(403, &gin.H{
			"code":  0,
			"error": "Your token has insufficient scope",
		})
		return
	}

	// Decode the input
	var input struct {
		Meta map[string]interface{} `json:"meta"`
		Body []byte                 `json:"body"`
		Tags []string               `json:"tags"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	// Save it into database
	resource := &models.Resource{
		ID:           uniuri.NewLen(uniuri.UUIDLen),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        account.ID,

		Meta: input.Meta,
		Body: input.Body,
		Tags: input.Tags,
	}

	// Insert it into database
	if err := r.Table("resources").Insert(resource).Exec(a.Rethink); err != nil {
		c.JSON(500, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	c.JSON(201, resource)
}

func (a *API) getAccountResources(c *gin.Context) {
	// Token and account from context
	var (
		ownAccount = c.MustGet("account").(*models.Account)
		token      = c.MustGet("token").(*models.Token)
	)

	// Resolve the ID from the URL
	id := c.Param("id")
	if id == "me" {
		id = ownAccount.ID
	}

	// Check the scope
	if id == ownAccount.ID {
		if !models.InScope(token.Scope, []string{"resources:read"}) {
			c.JSON(403, &gin.H{
				"code":  0,
				"error": "Your token has insufficient scope",
			})
			return
		}
	} else {
		if !models.InScope(token.Scope, []string{"admin"}) {
			c.JSON(403, &gin.H{
				"code":  0,
				"error": "Your token has insufficient scope",
			})
			return
		}
	}

	// Get resources from database without bodies
	cursor, err := r.Table("resources").GetAllByIndex("owner", id).Without("body").Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	defer cursor.Close()
	var resources []*models.Resource
	if err := cursor.All(&resources); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	// Write the response
	c.JSON(200, resources)
	return
}
