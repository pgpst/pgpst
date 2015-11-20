package config

type Config struct {
	// General settings
	LogLevel      string `default:"debug"`
	DefaultDomain string
	// Server-specific settings
	API    APIConfig
	Mailer MailerConfig
	// Metadata database settings
	Database  Database
	RethinkDB RethinkDBConfig
	Tiedot    TiedotConfig
	// Blob storage settings
	Storage    Storage
	WeedFS     WeedFSConfig
	Filesystem FilesystemConfig
	// Message queue settings
	Queue  Queue
	NSQ    NSQConfig
	Memory MemoryConfig
}

// API settings
type APIConfig struct {
	Enabled bool
	Address string
}

// Mailer settings
type MailerConfig struct {
	Enabled           bool
	Address           string
	SenderConcurrency int
	TLSCert           string
	TLSKey            string
	WelcomeMessage    string
	ReadTimeout       int
	WriteTimeout      int
	DataTimeout       int
	MaxConnections    int
	MaxMessageSize    int
	MaxRecipients     int
	RelayAddress      string
	SpamdAddress      string
}

// Database types
type Database int

const (
	RethinkDB Database = iota
	Tiedot
)

// RethinkDB settings
type RethinkDBConfig struct {
	Addresses     []string
	Database      string
	AuthKey       string
	Timeout       int
	WriteTimeout  int
	ReadTimeout   int
	MaxIdle       int
	MaxOpen       int
	DiscoverHosts bool
}

// Tiedot settings
type TiedotConfig struct {
	Path string
}

// Storage types
type Storage int

const (
	WeedFS Storage = iota
	Filesystem
)

// WeedFS configuration
type WeedFSConfig struct {
	MasterURL string
}

// Filesystem configuration
type FilesystemConfig struct {
	Path string
}

// Queue types
type Queue int

const (
	NSQ Queue = iota
	Memory
)

type NSQConfig struct {
	ServerAddress  string
	LookupdAddress string
}

type MemoryConfig struct{}
