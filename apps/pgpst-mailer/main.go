package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pgpst/pgpst/internal/github.com/namsral/flag"

	"github.com/pgpst/pgpst/pkg/mailer"
)

func mailerFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("mailer", flag.ExitOnError)

	// Basic options
	fs.String("config", "", "path to the config file")
	fs.Bool("version", false, "print version of the api")
	fs.String("log_level", "info", "lowest level of log messages to print")

	// RethinkDB connection
	fs.String("rethinkdb", "rethinkdb://127.0.0.1:28015/prod", "RethinkDB connection string")

	// NSQ connection
	fs.String("nsqd_address", "127.0.0.1:4150", "Address of the nsqd server to use")
	fs.String("lookupd_address", "127.0.0.1:4160", "Address of the lookupd server to use")

	// Error reporting
	fs.String("raven_dsn", "", "SQN to use by Raven, client of Sentry")

	// Server settings
	fs.String("smtp_address", "0.0.0.0:25", "Address of the SMTP server")
	fs.Bool("smtp_force_tls", false, "Force TLS?")
	fs.String("smtp_tls_cert", "", "Path to the TLS cert to use")
	fs.String("smtp_tls_key", "", "Path to the TLS key to use")

	// Connection parameters
	fs.String("hostname", "pgp.st", "Hostname of the mailer")
	fs.Int("read_timeout", 0, "Connection read timeout in seconds")
	fs.Int("write_timeout", 0, "Connection write timeout in seconds")
	fs.Int("data_timeout", 0, "Data read timeout in seconds")
	fs.Int("max_connections", 0, "Max concurrent connections")
	fs.Int("max_message_size", 0, "Max size of a single envelope in bytes")
	fs.Int("max_recipients", 0, "Max recipients per envelope")

	// Spamd address
	fs.String("spamd_address", "127.0.0.1:783", "Address of the spamd server to use")

	// Outbound email settings
	fs.String("smtpd_address", "127.0.0.1:2525", "Address of the SMTP relay to use")
	fs.Int("dkim_lru_size", 128, "Size of the in-memory DKIM key cache")
	fs.Int("sender_concurrency", 10, "Max concurrency of the email sender")

	return fs
}

func main() {
	// Parse flags
	fs := mailerFlagSet()
	fs.Parse(os.Args[1:])

	// Print the version if -version=true
	if fs.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("pgpst-mailer"))
		return
	}

	// Create a new signal receiver
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	// Create a new mailer
	mailer := mailer.NewMailer(mailer.NewOptions(fs))

	// Spawn it in a new goroutine
	go mailer.Main()

	// Watch for a signal
	<-sc

	// Exit the mailer
	mailer.Exit()
}
