package models

import (
	"time"
)

type Address struct {
	ID           string    `gorethink:"id"`                           // the actual address
	DateCreated  time.Time `gorethink:"date_created,omitempty"`       // when it was created
	DateModified time.Time `gorethink:"date_modified,omitempty"`      // last update
	Owner        string    `gorethink:"owner"`                        // who owns it
	PublicKey    string    `json:"public_key" gorethink:"public_key"` // default public key
}
