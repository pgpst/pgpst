package utils_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/utils"
)

func TestGinLogger(t *testing.T) {
	Convey("Given a buffered logrus", t, func() {
		log := logrus.New()
		buf := &bytes.Buffer{}
		log.Out = buf

		Convey("And a configured HTTP server", func() {
			gin.SetMode(gin.ReleaseMode)
			router := gin.New()

			router.Use(utils.GinLogger("API", log))

			router.GET("/", func(c *gin.Context) {
				c.JSON(200, "hello")
			})

			ts := httptest.NewServer(router)

			Convey("A message should be logged to the buffer upon a request", func() {
				resp, err := http.Get(ts.URL + "/")
				So(err, ShouldBeNil)

				body, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				So(string(body), ShouldEqual, "\"hello\"\n")

				So(buf.String(), ShouldContainSubstring, "API request finished")
				So(buf.String(), ShouldContainSubstring, "method=GET path=\"/\" status=200")
			})
		})
	})
}
