package api

import (
	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	"github.com/pgpst/pgpst/internal/github.com/namsral/flag"
)

type Options struct {
	LogLevel logrus.Level

	DefaultDomain string
	HTTPAddress   string

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

		DefaultDomain: fs.Lookup("default_domain").Value.String(),
		HTTPAddress:   fs.Lookup("http_address").Value.String(),

		RethinkDBAddress:  fs.Lookup("rethinkdb_address").Value.String(),
		RethinkDBDatabase: fs.Lookup("rethinkdb_database").Value.String(),

		NSQdAddress:    fs.Lookup("nsqd_address").Value.String(),
		LookupdAddress: fs.Lookup("lookupd_address").Value.String(),

		RavenDSN: fs.Lookup("raven_dsn").Value.String(),

		YubiCloudID:  fs.Lookup("yubicloud_id").Value.String(),
		YubiCloudKey: fs.Lookup("yubicloud_key").Value.String(),
	}
}
