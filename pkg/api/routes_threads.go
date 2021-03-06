package api

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

type extendedThread struct {
	*models.Thread
	Manifest []byte `gorethink:"manifest" json:"manifest"`
}

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
	cursor, err = r.Table("threads").GetAllByIndex("labels", label.ID).OrderBy(r.Desc("date_modified")).Map(func(thread r.Term) r.Term {
		return thread.Merge(map[string]interface{}{
			"manifest": r.Table("emails").GetAllByIndex("thread", thread.Field("id")).OrderBy("date_modified").CoerceTo("array"),
		}).Do(func(thread r.Term) r.Term {
			return r.Branch(
				thread.Field("manifest").Count().Gt(0),
				thread.Merge(map[string]interface{}{
					"manifest": thread.Field("manifest").Nth(0).Field("manifest"),
				}),
				thread.Without("manifest"),
			)
		})
	}).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	var threads []*extendedThread
	if err := cursor.All(&threads); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	if threads == nil {
		threads = []*extendedThread{}
	}

	// Write the response
	c.JSON(200, threads)
	return
}
