package utils_test

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/utils"
)

func TestCLIRecovery(t *testing.T) {
	Convey("Given a buffer", t, func() {
		buf := &bytes.Buffer{}

		Convey("CLIRecovery should write the stack to the buffer properly", func() {
			Convey("With error panic", func() {
				So(func() {
					defer utils.CLIRecovery(buf)

					panic(errors.New("hello"))
				}, ShouldNotPanic)

				So(buf.String(), ShouldContainSubstring, "cli_recovery_test.go:22")
			})

			Convey("With string panic", func() {
				So(func() {
					defer utils.CLIRecovery(buf)

					panic("hello")
				}, ShouldNotPanic)

				So(buf.String(), ShouldContainSubstring, "cli_recovery_test.go:32")
			})
		})
	})
}
