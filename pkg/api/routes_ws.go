package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) ws(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"hey": "there",
	})
}
