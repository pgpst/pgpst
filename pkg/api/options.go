package api

import (
	"flag"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
)

type Options struct {
	LogLevel logrus.Level

	HTTPAddress         string
	SessionDuration     time.Duration
	SessionDurationLong time.Duration

	RethinkDBAddress  string
	RethinkDBDatabase string

	NSQdAddress    string
	LookupdAddress string

	RavenDSN string

	YubiCloudID  string
	YubiCloudKey string
}

var llMapping = map[string]logrus.Level{
	"panic": logrus.PanicLevel,
	"fatal": logrus.FatalLevel,
	"error": logrus.ErrorLevel,
	"warn":  logrus.WarnLevel,
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
}

func NewOptions(fs *flag.FlagSet) *Options {
	ll, ok := llMapping[fs.Lookup("log_level").Value.String()]
	if !ok {
		ll = logrus.InfoLevel
	}

	return &Options{
		LogLevel: ll,

		HTTPAddress: fs.Lookup("http_address").Value.String(),
		SessionDuration: time.Duration(
			fs.Lookup("session_duration").Value.(flag.Getter).Get().(int)) * time.Hour,
		SessionDurationLong: time.Duration(
			fs.Lookup("session_duration_long").Value.(flag.Getter).Get().(int)) * time.Hour,

		RethinkDBAddress:  fs.Lookup("rethinkdb_address").Value.String(),
		RethinkDBDatabase: fs.Lookup("rethinkdb_database").Value.String(),

		NSQdAddress:    fs.Lookup("nsqd_address").Value.String(),
		LookupdAddress: fs.Lookup("lookupd_address").Value.String(),

		RavenDSN: fs.Lookup("raven_dsn").Value.String(),

		YubiCloudID:  fs.Lookup("yubicloud_id").Value.String(),
		YubiCloudKey: fs.Lookup("yubicloud_key").Value.String(),
	}
}
