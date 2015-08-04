package models

import (
	"time"
)

type Address struct {
	ID           string    `json:"id" gorethink:"id"`                                 // the actual address
	DateCreated  time.Time `json:"date_created" gorethink:"date_created,omitempty"`   // when it was created
	DateModified time.Time `json:"date_modified" gorethink:"date_modified,omitempty"` // last update
	Owner        string    `json:"owner" gorethink:"owner"`                           // who owns it
	PublicKey    string    `json:"public_key" gorethink:"public_key"`                 // default public key
}
