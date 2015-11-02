package api

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) getAccountAddresses(c *gin.Context) {
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
		if !models.InScope(token.Scope, []string{"addresses:read"}) {
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

	// Get addresses from database
	cursor, err := r.Table("addresses").GetAllByIndex("owner", id).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	defer cursor.Close()
	var addresses []*models.Address
	if err := cursor.All(&addresses); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	// Write the response
	c.JSON(200, addresses)
	return
}
