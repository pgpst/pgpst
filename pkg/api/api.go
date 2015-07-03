package api

import (
	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	"github.com/pgpst/pgpst/internal/github.com/bitly/go-nsq"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/getsentry/raven-go"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/utils"
)

type API struct {
	Options *Options

	Log      *logrus.Logger
	Rethink  *r.Session
	Producer *nsq.Producer
	Raven    *raven.Client
}

func NewAPI(options *Options) *API {
	// Create a new logger
	log := logrus.New()
	log.Level = options.LogLevel

	// Connect to the database
	session, err := r.Connect(r.ConnectOpts{
		Address:  options.RethinkDBAddress,
		Database: options.RethinkDBDatabase,
	})
	if err != nil {
		log.WithField("err", err).Fatal("Unable to connect to RethinkDB")
	}

	// Create a new NSQ producer
	producer, err := nsq.NewProducer(options.NSQdAddress, nsq.NewConfig())
	if err != nil {
		log.WithField("err", err).Fatal("Unable to connect to NSQd")
	}
	producer.SetLogger(&utils.NSQLogger{
		Log: log,
	}, nsq.LogLevelWarning)

	// Connect to Raven
	var rc *raven.Client
	if options.RavenDSN != "" {
		rc, err = raven.NewClient(options.RavenDSN, nil)
		if err != nil {
			log.WithField("err", err).Fatal("Unable to connect to Sentry")
		}
	}

	// Return a new API struct
	return &API{
		Options:  options,
		Log:      log,
		Rethink:  session,
		Producer: producer,
		Raven:    rc,
	}
}

func (a *API) Main() {
	// Create a new router
	router := gin.New()

	// Add two global middlewares
	router.Use(utils.GinLogger("API", a.Log))
	router.Use(utils.GinRecovery(a.Raven))

	router.GET("/")

	// Start the server
	if err := router.Run(a.Options.HTTPAddress); err != nil {
		a.Log.WithFields(logrus.Fields{
			"err":     err,
			"address": a.Options.HTTPAddress,
		}).Fatal("Unable to start a HTTP server")
	}
}

func (a *API) Exit() {

}
