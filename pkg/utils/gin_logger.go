package utils

import (
	"time"

	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"
)

func GinLogger(component string, log *logrus.Logger) gin.HandlerFunc {
	msg := component + " request finished"

	return func(c *gin.Context) {
		// Start a timer
		start := time.Now()

		// Process the request
		c.Next()

		// Stop the timer
		end := time.Now()
		latency := end.Sub(start)

		// Acquire info from the request
		fields := logrus.Fields{
			"ip":      c.ClientIP(),
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"status":  c.Writer.Status(),
			"latency": latency,
		}
		if x := c.Errors.ByType(gin.ErrorTypePrivate).String(); x != "" {
			fields["comment"] = x
		}

		// Write it to the logger
		log.WithFields(fields).Info(msg)
	}
}
