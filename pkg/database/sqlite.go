package database

import (
	"database/sql"
	"os/user"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"code.pgp.st/pgpst/pkg/config"
)

type SQLite struct {
	DB *sql.DB
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

	db, err := sql.Open("sqlite3", cfg.ConnectionString)
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
