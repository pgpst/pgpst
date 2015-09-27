package api

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) getLabelThreads(c *gin.Context) {
	// Token and account from context
	var (
		account = c.MustGet("account").(*models.Account)
		token   = c.MustGet("token").(*models.Token)
	)

	// Resolve the ID from the URL
	id := c.Param("id")

	// Get label from the database
	cursor, err := r.Table("labels").Get(id).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	var label *models.Label
	if err := cursor.One(&label); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	// Check the ownership and scope
	if label.Owner == account.ID {
		if !models.InScope(token.Scope, []string{"labels:read"}) {
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

	// Get threads from the database
	cursor, err = r.Table("threads").GetAllByIndex("labels", label.ID).OrderBy(r.Desc("date_modified")).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	var threads []*models.Thread
	if err := cursor.One(&threads); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	// Write the response
	c.JSON(200, threads)
	return
}
