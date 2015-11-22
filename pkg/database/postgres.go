package database

import (
	"database/sql"

	"code.pgp.st/pgpst/pkg/config"
	_ "github.com/lib/pq"
)

type Postgres struct {
	DB *sql.DB
}

func NewPostgres(cfg config.PostgresConfig) (*Postgres, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString)
	if err != nil {
		return nil, err
	}

	// Return the initialized struct
	return &Postgres{
		DB: db,
	}, nil
}

func (p *Postgres) Revision() (int, error) {
	var exists bool
	if err := p.DB.QueryRow(
		`SELECT to_regclass('migration_status') IS NOT NULL AS exists`,
	).Scan(&exists); err != nil {
		return 0, err
	}

	if exists {
		var revision int
		if err := p.DB.QueryRow(
			`SELECT value FROM migration_status WHERE key = 'revision'`,
		).Scan(&revision); err != nil {
			return 0, err
		}

		return revision, nil
	}

	if _, err := p.DB.Exec(
		`CREATE TABLE migration_status (
				key   text primary key,
				value int
			)`,
	); err != nil {
		return 0, err
	}
	if _, err := p.DB.Exec(
		`INSERT INTO migration_status VALUES('revision', -1)`,
	); err != nil {
		return 0, err
	}

	return -1, nil
}
