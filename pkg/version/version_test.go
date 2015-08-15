package version_test

import (
	"runtime"
	"testing"

	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/version"
)

func TestVersionFormat(t *testing.T) {
	Convey("version.String() should return a properly formatted string", t, func() {
		So(version.String("test"), ShouldEqual, "test v"+version.Version+" ("+runtime.Version()+")")
	})
}
