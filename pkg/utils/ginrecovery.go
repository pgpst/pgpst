package utils

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/pgpst/pgpst/internal/github.com/getsentry/raven-go"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"
)

func GinRecovery(rc *raven.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			var packet *raven.Packet
			switch rval := recover().(type) {
			case nil:
				return
			case error:
				packet = raven.NewPacket(
					rval.Error(),
					raven.NewHttp(c.Request),
					raven.NewException(rval, raven.NewStacktrace(2, 3, nil)),
				)
			default:
				str := fmt.Sprintf("%+v", rval)
				packet = raven.NewPacket(
					str,
					raven.NewHttp(c.Request),
					raven.NewException(errors.New(str), raven.NewStacktrace(2, 3, nil)),
				)
			}

			debug.PrintStack()
			c.AbortWithStatus(500)
			rc.Capture(packet, nil)
		}()

		c.Next()
	}
}
