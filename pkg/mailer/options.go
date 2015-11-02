package mailer

import (
	"strconv"

	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/namsral/flag"

	"github.com/pgpst/pgpst/pkg/utils"
)

type Options struct {
	LogLevel          logrus.Level
	RethinkOpts       r.ConnectOpts
	NSQdAddress       string
	LookupdAddress    string
	SenderConcurrency int
	SpamdAddress      string
	RavenDSN          string
	SMTPTLSCert       string
	SMTPTLSKey        string
	Hostname          string
	WelcomeMessage    string
	ReadTimeout       int
	WriteTimeout      int
	DataTimeout       int
	MaxConnections    int
	MaxMessageSize    int
	MaxRecipients     int
	SMTPAddress       string
	SMTPDAddress      string
}

var llMapping = map[string]logrus.Level{
	"panic": logrus.PanicLevel,
	"fatal": logrus.FatalLevel,
	"error": logrus.ErrorLevel,
	"warn":  logrus.WarnLevel,
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
}

func matoi(x int, y error) int {
	if y != nil {
		panic(y)
	}

	return x
}

func NewOptions(fs *flag.FlagSet) (*Options, error) {
	ll, ok := llMapping[fs.Lookup("log_level").Value.String()]
	if !ok {
		ll = logrus.InfoLevel
	}

	opts, err := utils.ParseRethinkDBString(fs.Lookup("rethinkdb").Value.String())
	if err != nil {
		return nil, err
	}

	return &Options{
		LogLevel:          ll,
		RethinkOpts:       opts,
		NSQdAddress:       fs.Lookup("nsqd_address").Value.String(),
		LookupdAddress:    fs.Lookup("lookupd_address").Value.String(),
		SenderConcurrency: matoi(strconv.Atoi(fs.Lookup("sender_concurrency").Value.String())),
		SpamdAddress:      fs.Lookup("spamd_address").Value.String(),
		RavenDSN:          fs.Lookup("raven_dsn").Value.String(),
		SMTPTLSCert:       fs.Lookup("smtp_tls_cert").Value.String(),
		SMTPTLSKey:        fs.Lookup("smtp_tls_key").Value.String(),
		Hostname:          fs.Lookup("hostname").Value.String(),
		WelcomeMessage:    fs.Lookup("welcome_message").Value.String(),
		ReadTimeout:       matoi(strconv.Atoi(fs.Lookup("read_timeout").Value.String())),
		WriteTimeout:      matoi(strconv.Atoi(fs.Lookup("write_timeout").Value.String())),
		DataTimeout:       matoi(strconv.Atoi(fs.Lookup("data_timeout").Value.String())),
		MaxConnections:    matoi(strconv.Atoi(fs.Lookup("max_connections").Value.String())),
		MaxMessageSize:    matoi(strconv.Atoi(fs.Lookup("max_message_size").Value.String())),
		SMTPAddress:       fs.Lookup("smtp_address").Value.String(),
		SMTPDAddress:      fs.Lookup("smtpd_address").Value.String(),
	}, nil
}
