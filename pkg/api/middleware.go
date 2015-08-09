package api

import (
	"strings"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) authMiddleware(c *gin.Context) {
	// Get the "Authorization" header
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		c.JSON(401, &gin.H{
			"code":  401,
			"error": "Invalid Authorization header",
		})
		c.Abort()
		return
	}

	// Split it into two parts - "Bearer" and token
	parts := strings.SplitN(authorization, " ", 2)
	if parts[0] != "Bearer" {
		c.JSON(401, &gin.H{
			"code":  401,
			"error": "Invalid Authorization header",
		})
		c.Abort()
		return
	}

	// Verify the token
	cursor, err := r.Table("tokens").Get(parts[1]).Do(func(token r.Term) map[string]interface{} {
		return map[string]interface{}{
			"token":   token,
			"account": r.Table("accounts").Get(token.Field("owner")),
		}
	}).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	defer cursor.Close()
	var result struct {
		Token   *models.Token   `gorethink:"token"`
		Account *models.Account `gorethink:"account"`
	}
	if err := cursor.One(&result); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		c.Abort()
		return
	}

	// Validate the token
	if result.Token.Type != "auth" {
		c.JSON(401, &gin.H{
			"code":  401,
			"error": "Invalid token type",
		})
		c.Abort()
		return
	}

	// Check for expiration
	if result.Token.IsExpired() {
		c.JSON(401, &gin.H{
			"code":  401,
			"error": "Your authentication token has expired",
		})
		c.Abort()
		return
	}

	// Validate the account
	if result.Account.Status != "active" {
		c.JSON(401, &gin.H{
			"code":  401,
			"error": "Your account is " + result.Account.Status,
		})
		c.Abort()
		return
	}

	// Write token into environment
	c.Set("account", result.Account)
	c.Set("token", result.Token)

	// Write some headers into the response
	c.Header(
		"X-Authenticated-As",
		result.Account.ID+"; "+result.Account.MainAddress,
	)
	c.Header(
		"X-Authenticated-Scope",
		strings.Join(result.Token.Scope, ", "),
	)
}
