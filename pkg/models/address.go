package models

import (
	"time"
)

type Address struct {
	ID           string    `db:"id"`
	StyledID     string    `db:"styled_id"`
	DateCreated  time.Time `db:"date_created"`
	DateModified time.Time `db:"date_modified"`
	Owner        string    `db:"owner"`
	PublicKey    string    `db:"public_key"`
}
