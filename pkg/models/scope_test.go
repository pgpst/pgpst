package models_test

import (
	"testing"

	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/models"
)

func TestScope(t *testing.T) {
	Convey("Given a scope and a comparsion slice that matches it", t, func() {
		scope := []string{
			"hello",
			"world",
		}

		what := []string{
			"hello:world",
			"world",
		}

		Convey("InScope should return true", func() {
			So(models.InScope(scope, what), ShouldBeTrue)
		})
	})

	Convey("Given a scope and a comparsion slice that doesn't match it", t, func() {
		scope := []string{
			"hello",
			"world",
		}

		what := []string{
			"foobar",
		}

		Convey("InScope should return false", func() {
			So(models.InScope(scope, what), ShouldBeFalse)
		})
	})
}
