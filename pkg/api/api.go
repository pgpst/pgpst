package api

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	log "gopkg.in/inconshreveable/log15.v2"

	"code.pgp.st/pgpst/pkg/config"
	"code.pgp.st/pgpst/pkg/database"
	"code.pgp.st/pgpst/pkg/queue"
	"code.pgp.st/pgpst/pkg/storage"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type API struct {
	Config   config.APIConfig
	Database database.Database
	Queue    queue.Queue
	Storage  storage.Storage

	Log    log.Logger
	Router *gin.Engine
}

func NewAPI(cf config.APIConfig, db database.Database, qu queue.Queue, st storage.Storage) (*API, error) {
	ap := &API{
		Config:   cf,
		Database: db,
		Queue:    qu,
		Storage:  st,
		Log:      log.New("module", "api"),
		Router:   gin.New(),
	}

	ap.Router.Use(CORS())
	ap.Router.Use(Logger(ap.Log))
	ap.Router.Use(Recovery())

	ap.Router.GET("/", ap.hello)
	ap.Router.Any("/ws", gin.WrapH(sockjs.NewHandler("/ws", sockjs.DefaultOptions, ap.ws)))

	return ap, nil
}

func (a *API) Start() error {
	a.Log.Info("Starting the API server", "address", a.Config.Address)

	if err := a.Router.Run(a.Config.Address); err != nil {
		a.Log.Crit("Unable to start the HTTP server", "error", err)
		return err
	}

	return nil
}
