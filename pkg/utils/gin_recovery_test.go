package utils_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pgpst/pgpst/internal/github.com/getsentry/raven-go"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"
	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/utils"
)

type grTransport struct {
	Output io.Writer
}

func (g *grTransport) Send(url, authHeader string, packet *raven.Packet) error {
	fmt.Fprintf(g.Output, "%s\n%s\n%+v\n", url, authHeader, packet)
	return nil
}

func TestGinRecovery(t *testing.T) {
	Convey("Given a configured HTTP server", t, func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()

		var rc *raven.Client
		router.Use(utils.GinRecovery(rc))

		router.GET("/test1", func(c *gin.Context) {
			panic("hello1")
		})
		router.GET("/test2", func(c *gin.Context) {
			panic(errors.New("hello2"))
		})
		router.GET("/test3", func(c *gin.Context) {
			c.JSON(200, "hello3")
		})

		ts := httptest.NewServer(router)

		Convey("String-valued panic should report to Sentry", func() {
			resp, err := http.Get(ts.URL + "/test1")
			So(err, ShouldBeNil)

			So(resp.StatusCode, ShouldEqual, 500)

			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldEqual, "{\"code\":500,\"error\":\"Internal server error\"}\n")
		})

		Convey("Error-valued panic should report to Sentry", func() {
			resp, err := http.Get(ts.URL + "/test2")
			So(err, ShouldBeNil)

			So(resp.StatusCode, ShouldEqual, 500)

			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldEqual, "{\"code\":500,\"error\":\"Internal server error\"}\n")
		})

		Convey("No panic should not report anything", func() {
			resp, err := http.Get(ts.URL + "/test3")
			So(err, ShouldBeNil)

			So(resp.StatusCode, ShouldEqual, 200)

			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldEqual, "\"hello3\"\n")

		})
	})
}
