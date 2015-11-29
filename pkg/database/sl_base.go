package database

import (
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"code.pgp.st/pgpst/pkg/config"
)

type SQLite struct {
	DB *sqlx.DB
}

func NewSQLite(cfg config.SQLiteConfig) (*SQLite, error) {
	if strings.Index(cfg.ConnectionString, "~") != -1 {
		parts := strings.Split(cfg.ConnectionString, "?")
		if strings.Index(parts[0], "~") == 0 {
			user, err := user.Current()
			if err != nil {
				return nil, err
			}

			parts[0] = filepath.Join(user.HomeDir, parts[0][1:])
			cfg.ConnectionString = strings.Join(parts, "?")
		}
	}

	db, err := sqlx.Connect("sqlite3", cfg.ConnectionString)
	if err != nil {
		return nil, err
	}

	// Return the initialized struct
	return &SQLite{
		DB: db,
	}, nil
}

func (s *SQLite) Revision() (int, error) {
	var exists bool
	if err := s.DB.QueryRow(
		`SELECT count(*) > 0 FROM sqlite_master WHERE type='table' AND name='migration_status'`,
	).Scan(&exists); err != nil {
		return 0, err
	}

	if exists {
		var revision int
		if err := s.DB.QueryRow(
			`SELECT value FROM migration_status WHERE key = 'revision'`,
		).Scan(&revision); err != nil {
			return 0, err
		}

		return revision, nil
	}

	if _, err := s.DB.Exec(
		`CREATE TABLE migration_status (
        	key   text primary key,
        	value int
        )`,
	); err != nil {
		return 0, err
	}

	if _, err := s.DB.Exec(
		`INSERT INTO migration_status VALUES('revision', -1)`,
	); err != nil {
		return 0, err
	}

	return -1, nil
}

func (s *SQLite) Migrate(id int) error {
	return Migrations[id].Migrate[config.SQLite](s.DB)
}

func (s *SQLite) SetRevision(id int) error {
	if _, err := s.DB.Exec(
		`UPDATE migration_status SET value = ? WHERE key = 'revision'`,
		id,
	); err != nil {
		return err
	}

	return nil
}
