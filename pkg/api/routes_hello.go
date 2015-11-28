package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code.pgp.st/pgpst/pkg/version"
)

func (a *API) hello(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"name": version.String("pgpst-api"),
	})
}
