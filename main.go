package main

import (
	"os"

	"github.com/koding/multiconfig"
	log "gopkg.in/inconshreveable/log15.v2"

	"code.pgp.st/pgpst/pkg/config"
	"code.pgp.st/pgpst/pkg/database"
	"code.pgp.st/pgpst/pkg/queue"
	"code.pgp.st/pgpst/pkg/storage"
)

func main() {
	// Load the configuration
	m := multiconfig.NewWithPath(os.Getenv("config"))
	m.Validator = &config.Validator{}
	cfg := &config.Config{}
	m.MustLoad(cfg)

	// Parse the log level
	lvl, err := log.LvlFromString(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	// Create a new logger
	l := log.New()
	l.SetHandler(
		log.LvlFilterHandler(lvl, log.StdoutHandler),
	)

	// Start the initialization
	il := l.New("module", "init")
	il.Info("Starting up the application")

	// Initialize the database
	var db database.Database
	switch cfg.Database {
	case config.Postgres:
		pdb, err := database.NewPostgres(cfg.Postgres)
		if err != nil {
			il.Crit("Unable to connect to the PostgreSQL server", "error", err)
			return
		}

		il.Info("Connected to a PostgreSQL server", "cstr", cfg.Postgres.ConnectionString)
		db = pdb
	case config.SQLite:
		sdb, err := database.NewSQLite(cfg.SQLite)
		if err != nil {
			il.Crit("Unable to load the SQLite3 database", "error", err)
			return
		}

		il.Info("Loaded a SQLite3 database", "cstr", cfg.SQLite.ConnectionString)
		db = sdb
	}

	// Get database's revision
	rev, err := db.Revision()
	if err != nil {
		il.Error("Unable to fetch database's revision", "error", err)
		return
	}
	il.Debug("Fetched the database revision", "rev", rev)

	// Run the migrations
	if cfg.Migrate && rev < len(database.Migrations)-1 {
		ml := l.New("module", "migration")

		for i := rev + 1; i < len(database.Migrations); i++ {
			if err := db.Migrate(i); err != nil {
				ml.Crit("Migration failed", "name", database.Migrations[i].Name, "error", err)
				return
			}

			if err := db.SetRevision(i); err != nil {
				ml.Crit("Unable to set the revision", "new", i, "error", err)
				return
			}

			ml.Info("Executed a migration", "index", i, "name", database.Migrations[i].Name)
		}

		ml.Info("Migration execution complete.")
	}

	// Load the storage engine
	var st storage.Storage
	switch cfg.Storage {
	case config.WeedFS:
		wst, err := storage.NewWeedFS(cfg.WeedFS)
		if err != nil {
			il.Crit("Unable to connect to the WeedFS storage", "error", err)
			return
		}

		il.Info("Connected to the WeedFS storage", "address", cfg.WeedFS.MasterURL)
		st = wst
	case config.Filesystem:
		fst, err := storage.NewFilesystem(cfg.Filesystem)
		if err != nil {
			il.Crit("Unable to load a filesystem storage", "error", err)
			return
		}

		il.Info("Loaded filesystem storage", "path", cfg.Filesystem.Path)
		st = fst
	}

	// Set up the queue
	var qu queue.Queue
	switch cfg.Queue {
	case config.NSQ:
		nqu, err := queue.NewNSQ(cfg.NSQ)
		if err != nil {
			il.Crit("Unable to connect to the queue", "error", err)
			return
		}

		il.Info("Connected to the NSQ queue")
		qu = nqu
	case config.Memory:
		mqu, err := queue.NewMemory()
		if err != nil {
			il.Crit("Unable to connect to the queue", "error", err)
			return
		}

		il.Info("Configured an in-memory queue")
		qu = mqu
	}

	_ = st
	_ = qu
}
