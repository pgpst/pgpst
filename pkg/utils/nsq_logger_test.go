package utils_test

import (
	"bytes"
	"testing"

	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/utils"
)

func TestNSQLogger(t *testing.T) {
	Convey("Given a buffered logrus", t, func() {
		log := logrus.New()
		buf := &bytes.Buffer{}
		log.Out = buf

		Convey("NSQLogger should print proper messages", func() {
			nsql := &utils.NSQLogger{
				Log: log,
			}

			So(nsql.Output(1, "Hello world"), ShouldBeNil)
			So(buf.String(), ShouldContainSubstring, "nsq_logger_test.go:24")
		})
	})
}
