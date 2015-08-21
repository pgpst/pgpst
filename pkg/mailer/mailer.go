package mailer

import (
	"crypto/tls"

	"github.com/hashicorp/lru"
	"github.com/lavab/go-spamc"
	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	"github.com/pgpst/pgpst/internal/github.com/bitly/go-nsq"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/getsentry/raven-go"
	"github.com/pgpst/smtpd"

	"github.com/pgpst/pgpst/pkg/utils"
)

type Mailer struct {
	Options *Options

	Log       *logrus.Logger
	Rethink   *r.Session
	Producer  *nsq.Producer
	Consumer  *nsq.Consumer
	Raven     *raven.Client
	Spam      *spamc.Client
	TLSConfig *tls.Config
}

func NewMailer(options *Options) *Mailer {
	// Create a new logger
	log := logrus.New()
	log.Level = options.LogLevel

	// Connect to the database
	session, err := r.Connect(options.RethinkConnectOpts)
	if err != nil {
		log.WithField("err", err).Fatal("Unable to connect to RethinkDB")
	}

	// Create a new NSQ producer
	producer, err := nsq.NewProducer(options.NSQdAddress, nsq.NewConfig())
	if err != nil {
		log.WithField("err", err).Fatal("Unable to connect to NSQd")
	}
	nsqlog := &utils.NSQLogger{
		Log: log,
	}
	producer.SetLogger(nsqlog, nsq.LogLevelWarning)

	// Prepare the struct
	mailer := &Mailer{
		Options:  options,
		Log:      log,
		Rethink:  session,
		Producer: producer,
	}

	// And a new NSQ consumer
	consumer, err := nsq.NewConsumer("send_email", "receive", nsq.NewConfig())
	if err != nil {
		log.WithField("err", err).Fatal("Unable to create a new NSQ consumer")
	}
	consumer.SetLogger(nsqlog, nsq.LogLevelWarning)
	consumer.AddConcurrentHandlers(mailer, options.SenderConcurrency)
	mailer.Consumer = consumer

	// Connect to spamd
	mailer.Spam = spamc.New(options.SpamdAddress, 10)

	// Connect to Raven
	var rc *raven.Client
	if options.RavenDSN != "" {
		rc, err = raven.NewClient(options.RavenDSN, nil)
		if err != nil {
			log.WithField("err", err).Fatal("Unable to connect to Sentry")
		}
	}
	mailer.Raven = rc

	// Load TLS config
	if options.SMTPTLSCert != "" {
		cert, err := ioutil.ReadFile(options.SMTPTLSCert)
		if err != nil {
			log.WithField("err", err).Fatal("Unable to read tls cert file")
		}

		key, err := ioutil.ReadFile(options.SMTPTLSKey)
		if err != nil {
			log.WithField("err", err).Fatal("Unable to read tls key file")
		}

		pair, err := tls.X509KeyPair(cert, key)
		if err != nil {
			log.WithField("err", err).Fatal("Unable to parse tls keypair")
		}

		mailer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{pair},
		}
	}

	// Return a new mailer struct
	return mailer
}

func (m *Mailer) Main() {
	// Create a handler
	smtp := &smtpd.Server{
		Hostname:       m.Options.Hostname,
		WelcomeMessage: m.Options.WelcomeMessage,

		ReadTimeout:  time.Second * time.Duration(m.Options.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(m.Options.WriteTimeout),
		DataTimeout:  time.Second * time.Duration(m.Options.DataTimeout),

		MaxConnections: m.Options.MaxConnections,
		MaxMessageSize: m.Options.MaxMessageSize,
		MaxRecipients:  m.Options.MaxRecipients,

		WrapperChain: []smtpd.Wrapper{
			m,
		},
		RecipientChain: []smtpd.Recipient{
			m,
		},
		DeliveryChain: []smtpd.Delivery{
			m,
		},

		TLSConfig: m.TLSConfig,
	}

	// Listen
	if err := smtp.ListenAndServe(m.Options.SMTPAddress); err != nil {
		log.WithField("err", err).Fatal("Unable to listen and serve the SMTP server")
	}
}

func (a *API) Exit() {

}
