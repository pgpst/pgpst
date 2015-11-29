package models

import (
	"time"
)

type Label struct {
	ID           string    `db:"id"`
	DateCreated  time.Time `db:"date_created"`
	DateModified time.Time `db:"date_modified"`
	Owner        string    `db:"owner"`
	Name         string    `db:"name"`
	IsSystem     bool      `db:"is_system"`
}
