package models

import (
	"time"
)

type Email struct {
	ID           string    `db:"id"`
	DateCreated  time.Time `db:"date_created"`
	DateModified time.Time `db:"date_modified"`
	Owner        string    `db:"owner"`
	MessageID    string    `db:"message_id"`
	Thread       string    `db:"thread"`
	Status       string    `db:"status"`
	PublicKey    string    `db:"public_key"`
	Manifest     []byte    `db:"manifest"`
	Body         string    `db:"body"`
}
