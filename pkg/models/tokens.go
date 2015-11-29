package models

import (
	"time"
)

type Thread struct {
	ID           string    `db:"id"`
	DateCreated  time.Time `db:"date_created"`
	DateModified time.Time `db:"date_modified"`
	Owner        string    `db:"owner"`
	ExpiryDate   time.Time `db:"expiry_date"`
	Type         string    `db:"type"`
	SScope       string    `db:"scope"`
	Scope        []string  `db:"-"`
	AppID        string    `db:"app_id"`
}
