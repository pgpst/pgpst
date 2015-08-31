package api

import (
	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	"github.com/pgpst/pgpst/internal/github.com/bitly/go-nsq"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/getsentry/raven-go"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/utils"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

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
	router.Use(utils.GinCORS())
	router.Use(utils.GinLogger("API", a.Log))
	router.Use(utils.GinRecovery(a.Raven))

	// Hello route
	router.GET("/", a.hello)

	v1 := router.Group("/v1")
	{
		// Public routes
		v1.POST("/accounts", a.createAccount) // Registration and reservation
		v1.POST("/oauth", a.oauthToken)       // Various OAuth handlers
		v1.GET("/keys/:id", a.readKey)        // Open keyserver

		// Create a subrouter
		v1a := v1.Group("/", a.authMiddleware)
		{
			// Accounts
			//v1a.GET("/accounts", a.listAccounts)
			v1a.GET("/accounts/:id", a.readAccount)
			v1a.PUT("/accounts/:id", a.updateAccount)
			//v1a.DELETE("/accounts/:id", a.deleteAccount)
			//v1a.GET("/accounts/:id/keys", a.getAccountKeys)

			// Addresses
			//v1a.POST("/addresses", a.createAddress)
			//v1a.GET("/addresses", a.listAddresses)
			//v1a.GET("/addresses/:id", a.readAddress)
			//v1a.PUT("/addresses/:id", a.updateAddress)
			//v1a.DELETE("/addresses/:id", a.deleteAddress)

			// Emails
			//v1a.POST("/emails", a.createEmail)
			//v1a.GET("/emails", a.listEmails)
			//v1a.GET("/emails/:id", a.getEmail)
			//v1a.DELETE("/emails/:id", a.deleteEmail)

			// Keys
			v1a.POST("/keys", a.createKey)
			//v1a.GET("/keys", a.listKeys)
			//v1a.PUT("/keys/:id", a.updateKeys)
			//v1a.DELETE("/keys/:id", a.deleteKey)

			// Labels
			//v1a.POST("/labels", a.createLabel)
			//v1a.GET("/labels", a.listLabels)
			//v1a.GET("/labels/:id", a.readLabel)
			//v1a.PUT("/labels/:id", a.updateLabel)
			//v1a.DELETE("/labels/:id", a.deleteLabel)

			// Threads
			//v1a.GET("/threads", a.listThreads)
			//v1a.GET("/threads/:id", a.readThread)
			//v1a.PUT("/threads/:id", a.updateThread)
			//v1a.DELETE("/threads/:id", a.deleteThread)

			// Tokens
			v1a.POST("/tokens", a.createToken)
			//v1a.GET("/tokens", a.listTokens)
			//v1a.GET("/tokens/:id", a.readToken)
			//v1a.PUT("/tokens/:id", a.updateToken)
			//v1a.DELETE("/tokens/:id", a.deleteToken)

			// Resources
			//v1a.POST("/resources", a.createResource)
			//v1a.GET("/resources", a.listResources)
			//v1a.GET("/resources/:id", a.readResource)
			//v1a.PUT("/resources/:id", a.updateResource)
			//v1a.DELETE("/resources/:id", a.deleteResource)
		}
	}

	// Log that we're about to start the server
	a.Log.WithFields(logrus.Fields{
		"address": a.Options.HTTPAddress,
	}).Info("Starting the server")

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
