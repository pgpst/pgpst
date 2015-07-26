package models

import (
	"time"
)

type Key struct {
	ID           string    `json:"id" gorethink:"id"`                                           // key's fingerprint
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // when it was created
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // last update
	Owner        string    `json:"owner,omitempty" gorethink:"owner,omitempty"`                 // owner of the key

	ExpiryDate time.Time         `json:"expiry_date,omitempty" gorethink:"expiry_date,omitempty"`   // when it expires
	Headers    map[string]string `json:"headers,omitempty" gorethink:"headers,omitempty"`           // headers from the orig data
	Algorithm  string            `json:"algorithm,omitempty" gorethink:"algorithm,omitempty"`       // algorithm of the key
	Length     int               `json:"length,omitempty" gorethink:"length,omitempty"`             // key's length
	Key        []byte            `json:"key,omitempty" gorethink:"key,omitempty"`                   // the actual key
	KeyID      string            `json:"key_id,omitempty" gorethink:"key_id,omitempty"`             // key_id
	KeyIDShort string            `json:"key_id_short,omitempty" gorethink:"key_id_short,omitempty"` // shorter version of key_id
	MasterKey  string            `json:"master_key,omitempty" gorethink:"master_key,omitempty"`     // master key
}
