package models

import (
	"time"
)

type Resource struct {
	ID           string                 `db:"id"`
	DateCreated  time.Time              `db:"date_created"`
	DateModified time.Time              `db:"date_modified"`
	Owner        string                 `db:"owner"`
	Body         string                 `db:"body"`
	STags        string                 `db:"tags"`
	Tags         []string               `db:"-"`
	SMeta        string                 `db:"meta"`
	Meta         map[string]interface{} `db:"-"`
}
