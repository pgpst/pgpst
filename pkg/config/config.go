package config

type Config struct {
	// Migrate flag
	Migrate bool
	// General settings
	LogLevel      string `default:"debug"`
	DefaultDomain string `default:"pgp.st"`
	// Server-specific settings
	API    APIConfig
	Mailer MailerConfig
	// Metadata database settings
	Database Database `default:"sqlite"`
	Postgres PostgresConfig
	SQLite   SQLiteConfig
	// Blob storage settings
	Storage    Storage `default:"filesystem"`
	WeedFS     WeedFSConfig
	Filesystem FilesystemConfig
	// Message queue settings
	Queue  Queue `default:"memory"`
	NSQ    NSQConfig
	Memory MemoryConfig
}

// API settings
type APIConfig struct {
	Enabled bool   `default:"false"`
	Address string `default:"0.0.0.0:6030"`
}

// Mailer settings
type MailerConfig struct {
	Enabled           bool   `default:"false"`
	Address           string `default:"0.0.0.0:25"`
	SenderConcurrency int    `default:"10"`
	TLSCert           string
	TLSKey            string
	WelcomeMessage    string `default:"Welcome to the pgp.st Mailer"`
	ReadTimeout       int    `default:"60"`
	WriteTimeout      int    `default:"60"`
	DataTimeout       int    `default:"300"`
	MaxConnections    int    `default:"100"`
	MaxMessageSize    int    `default:"104857600"`
	MaxRecipients     int    `default:"100"`
}

// Database types
type Database string

const (
	Postgres Database = "postgres"
	SQLite            = "sqlite"
)

// Postgres settings
type PostgresConfig struct {
	ConnectionString string `default:"postgres://127.0.0.1:5432/pgpst"`
}

// Tiedot settings
type SQLiteConfig struct {
	ConnectionString string `default:"~/.pgpst/database.db"`
}

// Storage types
type Storage string

const (
	WeedFS     Storage = "weedfs"
	Filesystem         = "filesystem"
)

// WeedFS configuration
type WeedFSConfig struct {
	MasterURL string
}

// Filesystem configuration
type FilesystemConfig struct {
	Path string `default:"~/.pgpst/storage"`
}

// Queue types
type Queue string

const (
	NSQ    Queue = "nsq"
	Memory       = "memory"
)

type NSQConfig struct {
	ServerAddress  string
	LookupdAddress string
}

type MemoryConfig struct{}
