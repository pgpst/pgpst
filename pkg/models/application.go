package models

import (
	"time"
)

type Application struct {
	ID           string    `db:"id"`
	DateCreated  time.Time `db:"date_created"`
	DateModified time.Time `db:"date_modified"`
	Owner        string    `db:"owner"`
	Secret       string    `db:"secret"`
	Callback     string    `db:"callback"`
	Name         string    `db:"name"`
	HomePage     string    `db:"homepage"`
	Description  string    `db:"description"`
}
