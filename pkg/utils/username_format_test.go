package utils_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/utils"
)

func TestUsernameFormat(t *testing.T) {
	Convey("Given a funky-spelled email address", t, func() {
		address := "P.i.o.T+.r@pGP.St"

		Convey("NormalizeAddress should simplify it", func() {
			address = utils.NormalizeAddress(address)

			So(address, ShouldEqual, "p.i.o.t.r@pgp.st")

			Convey("And RemoveDots should remove all dots from the username part", func() {
				address = utils.RemoveDots(address)
				So(address, ShouldEqual, "piotr@pgp.st")
			})
		})
	})

	Convey("Given a string with dots", t, func() {
		input := "he.....llo."

		Convey("RemoveDots should remove all dots", func() {
			input = utils.RemoveDots(input)
			So(input, ShouldEqual, "hello")
		})
	})
}
