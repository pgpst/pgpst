package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"code.pgp.st/pgpst/pkg/config"
)

type Postgres struct {
	DB *sqlx.DB
}

func NewPostgres(cfg config.PostgresConfig) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", cfg.ConnectionString)
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

func (p *Postgres) Migrate(id int) error {
	return Migrations[id].Migrate[config.Postgres](p.DB)
}

func (p *Postgres) SetRevision(id int) error {
	if _, err := p.DB.Exec(
		`UPDATE migration_status SET value = $1 WHERE key = 'revision'`,
		id,
	); err != nil {
		return err
	}

	return nil
}
