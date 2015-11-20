package main

import (
	"os"

	"github.com/koding/multiconfig"
	log "gopkg.in/inconshreveable/log15.v2"

	"code.pgp.st/pgpst/pkg/config"
)

func main() {
	// Load the configuration
	m := multiconfig.NewWithPath(os.Getenv("config"))
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
}
