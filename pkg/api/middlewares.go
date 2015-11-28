package api

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "gopkg.in/inconshreveable/log15.v2"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// We will have a SockJS endpoint on /ws
		if strings.HasPrefix(c.Request.RequestURI, "/v1/ws") {
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

func Logger(lo log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a timer
		start := time.Now()

		// Process the request
		c.Next()

		// Stop the timer
		end := time.Now()
		latency := end.Sub(start)

		// Pass it to the logger
		log.Info(
			"HTTP request finished",
			"ip", c.ClientIP(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency,
		)
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			var err error

			switch rval := recover().(type) {
			case nil:
				return
			case error:
				err = rval
			default:
				err = fmt.Errorf("%+v", err)
			}

			debug.PrintStack()

			c.JSON(500, &gin.H{
				"code":  500,
				"error": "Internal server error",
			})
			c.Abort()
		}()

		c.Next()
	}
}
