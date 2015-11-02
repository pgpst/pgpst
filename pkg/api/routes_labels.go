package api

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) getAccountLabels(c *gin.Context) {
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

	// Get labels from database
	cursor, err := r.Table("labels").GetAllByIndex("owner", id).Map(func(label r.Term) r.Term {
		return label.Merge(map[string]interface{}{
			"total_threads": r.Table("threads").GetAllByIndex("labels", label.Field("id")).Count(),
			"unread_threads": r.Table("threads").GetAllByIndex("labelsIsRead", []interface{}{
				label.Field("id"),
				false,
			}).Count(),
		})
	}).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	defer cursor.Close()
	var labels []struct {
		models.Label
		TotalThreads  int `json:"total_threads" gorethink:"total_threads"`
		UnreadThreads int `json:"unread_threads" gorethink:"unread_threads"`
	}
	if err := cursor.All(&labels); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	// Write the response
	c.JSON(200, labels)
	return
}
