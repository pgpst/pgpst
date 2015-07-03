package api

import (
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"
	"net/http"

	"github.com/pgpst/pgpst/pkg/version"
)

func hello(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"name": version.String("pgpst-api"),
	})
}
