package utils

import (
	"strings"

	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"
)

func GinCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// We will have a SockJS endpoint on /ws
		if strings.HasPrefix(c.Request.RequestURI, "/ws") {
			c.Next()
			return
		}

		// Enable credentials
		c.Header("Access-Control-Allow-Credentials", "true")

		// Default headers
		allowedHeaders := []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Accept",
			"Cache-Control",
			"X-Requested-With",
		}

		// Expand them with requested headers
		requestedHeaders := strings.Split(c.Request.Header.Get("Access-Control-Request-Headers"), ",")
		allowedHeaders = append(allowedHeaders, requestedHeaders...)

		// Remove duplicates
		resultHeaders := []string{}
		seenHeaders := map[string]struct{}{}
		for _, header := range allowedHeaders {
			if _, ok := seenHeaders[header]; !ok && header != "" {
				resultHeaders = append(resultHeaders, header)
				seenHeaders[header] = struct{}{}
			}
		}

		// Set allow-headers
		c.Header("Access-Control-Allow-Headers", strings.Join(resultHeaders, ","))

		// Allow all methods and all origins
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS,DELETE")
		c.Header("Access-Control-Allow-Origin", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
