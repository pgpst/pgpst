package models

import (
	"time"
)

type Application struct {
	ID           string    `json:"id" gorethink:"id"`
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"`
	Owner        string    `json:"owner" gorethink:"owner"`

	Secret   string `json:"secret" gorethink:"secret"`
	Callback string `json:"callback" gorethink:"callback"`

	Logo        []byte `json:"logo" gorethink:"logo"`
	Homepage    string `json:"homepage" gorethink:"homepage"`
	Name        string `json:"name" gorethink:"name"`
	Description string `json:"description" gorethink:"description"`
}
