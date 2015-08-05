package utils_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/utils"
)

type dumbReader struct{}

func (d *dumbReader) Read(a []byte) (int, error) {
	return 0, errors.New("random error")
}

func TestCLIConfirmation(t *testing.T) {
	Convey("Given a buffer", t, func() {
		w := &bytes.Buffer{}
		r := &bytes.Buffer{}

		Convey("AskForConfirmation should respond correctly to all cases", func() {
			// Yes case
			var wg sync.WaitGroup
			wg.Add(1)

			var (
				resp bool
				cerr error
			)
			go func() {
				resp, cerr = utils.AskForConfirmation(w, r, "hello: ")
				wg.Done()
			}()

			r.WriteString("yes\n")
			wg.Wait()

			So(cerr, ShouldBeNil)
			So(w.String(), ShouldEqual, "hello: ")
			So(resp, ShouldBeTrue)

			// Clean up
			w.Reset()
			r.Reset()

			// No case
			wg.Add(1)
			go func() {
				resp, cerr = utils.AskForConfirmation(w, r, "hello: ")
				wg.Done()
			}()

			r.WriteString("no\n")
			wg.Wait()

			So(cerr, ShouldBeNil)
			So(w.String(), ShouldEqual, "hello: ")
			So(resp, ShouldBeFalse)

			// Clean up for the second time
			w.Reset()
			r.Reset()

			// Retry case
			wg.Add(1)
			go func() {
				resp, cerr = utils.AskForConfirmation(w, r, "hello: ")
				wg.Done()
			}()

			r.WriteString("invalid\n")
			r.WriteString("yes\n")
			wg.Wait()

			So(cerr, ShouldBeNil)
			So(w.String(), ShouldEqual, "hello: hello: ")
			So(resp, ShouldBeTrue)

			// Closed stdin case
			w.Reset()
			r2 := &dumbReader{}
			resp, err := utils.AskForConfirmation(w, r2, "hello: ")
			So(err, ShouldNotBeNil)
			So(w.String(), ShouldEqual, "hello: ")
			So(resp, ShouldBeFalse)
		})
	})
}
