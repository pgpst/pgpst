package models_test

import (
	"testing"
	"time"

	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/models"
)

func TestToken(t *testing.T) {
	Convey("Given an expired token", t, func() {
		token := &models.Token{
			ExpiryDate: time.Now().Truncate(time.Hour),
		}

		Convey("IsExpired should return true", func() {
			So(token.IsExpired(), ShouldBeTrue)
		})
	})

	Convey("Given a token that has not expired yet", t, func() {
		token := &models.Token{
			ExpiryDate: time.Now().Add(time.Hour),
		}

		Convey("IsExpired should return true", func() {
			So(token.IsExpired(), ShouldBeFalse)
		})
	})

	Convey("Given a token with no expiration", t, func() {
		token := &models.Token{}

		Convey("IsExpired should return true", func() {
			So(token.IsExpired(), ShouldBeFalse)
		})
	})
}
