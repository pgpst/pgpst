package models

import (
	"time"
)

type Account struct {
	ID           string    `db:"id"`
	DateCreated  time.Time `db:"date_created"`
	DateModified time.Time `db:"date_modified"`
	MainAddress  string    `db:"main_address"`
	Password     []byte    `db:"password"`
	Subscription string    `db:"subscription"`
	AltEmail     string    `db:"alt_email"`
	Status       string    `db:"status"`
}
