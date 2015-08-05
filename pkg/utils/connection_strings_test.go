package utils_test

import (
	"testing"
	"time"

	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/utils"
)

func TestRethinkDBConnectionString(t *testing.T) {
	Convey("Given an invalid connection string", t, func() {
		cs := "://123123zxc9123=-_+_+_+"

		Convey("Parsing should fail", func() {
			_, err := utils.ParseRethinkDBString(cs)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given a valid RethinkDB connection string", t, func() {
		cs := "rethinkdb://authkey@127.0.0.1:28015/prod?discover_hosts=true&refresh_interval=5&max_open=30&max_idle=10"
		Convey("Parsing should succeed", func() {
			opts, err := utils.ParseRethinkDBString(cs)
			So(err, ShouldBeNil)
			So(opts.Address, ShouldEqual, "127.0.0.1:28015")
			So(opts.AuthKey, ShouldEqual, "authkey")
			So(opts.Database, ShouldEqual, "prod")
			So(opts.DiscoverHosts, ShouldBeTrue)
			So(opts.NodeRefreshInterval, ShouldEqual, time.Second*5)
			So(opts.MaxOpen, ShouldEqual, 30)
			So(opts.MaxIdle, ShouldEqual, 10)
		})
	})

	Convey("Given a non-RethinkDB connection string", t, func() {
		cs := "http://google.com"
		Convey("Parsing should fail", func() {
			_, err := utils.ParseRethinkDBString(cs)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given RethinkDB connection strings with invalid params", t, func() {
		cs1 := "rethinkdb://127.0.0.1:28015/prod?refresh_interval=asd"
		cs2 := "rethinkdb://127.0.0.1:28015/prod?max_open=asd"
		cs3 := "rethinkdb://127.0.0.1:28015/prod?max_idle=asd"

		Convey("Parsing should fail each time", func() {
			_, err := utils.ParseRethinkDBString(cs1)
			So(err, ShouldNotBeNil)
			_, err = utils.ParseRethinkDBString(cs2)
			So(err, ShouldNotBeNil)
			_, err = utils.ParseRethinkDBString(cs3)
			So(err, ShouldNotBeNil)
		})
	})
}
