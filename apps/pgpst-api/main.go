package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pgpst/pgpst/internal/github.com/namsral/flag"

	"github.com/pgpst/pgpst/pkg/api"
	"github.com/pgpst/pgpst/pkg/version"
)

func apiFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("api", flag.ExitOnError)

	// Basic options
	//fs.String("config", "", "path to the config file")
	fs.Bool("version", false, "print version of the api")
	fs.String("log_level", "info", "lowest level of log messages to print")

	// App settings
	fs.String("default_domain", "pgp.st", "Default email domain")
	fs.String("http_address", "0.0.0.0:8000", "Address of the HTTP server")

	// RethinkDB connection
	fs.String("rethinkdb_address", "127.0.0.1:28015", "Address to the RethinkDB server")
	fs.String("rethinkdb_database", "prod", "Name of the database to use")

	// NSQ connection
	fs.String("nsqd_address", "127.0.0.1:4150", "Address of the nsqd server to use")
	fs.String("lookupd_address", "127.0.0.1:4160", "Address of the nsqlookupd server to use")

	// Error reporting
	fs.String("raven_dsn", "", "DSN to use by Raven, the client of Sentry")

	// Two-factor authentication options
	fs.String("yubicloud_id", "", "App ID for the YubiCloud API")
	fs.String("yubicloud_key", "", "Key for the YubiCloud API")

	return fs
}

func main() {
	// Parse flags
	fs := apiFlagSet()
	fs.Parse(os.Args[1:])

	// Print the version if -version=true
	if fs.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("pgpst-api"))
		return
	}

	// Create a new signal receiver
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	// Create a new API
	api := api.NewAPI(api.NewOptions(fs))

	// Spawn it in a new goroutine
	go api.Main()

	// Watch for a signal
	<-sc

	// Exit the API
	api.Exit()
}
