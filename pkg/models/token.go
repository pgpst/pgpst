package models

import (
	"time"
)

type Token struct {
	ID           string    `json:"id" gorethink:"id"`                                           // 20-char id
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // time of creation
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // time of last mod
	Owner        string    `json:"owner" gorethink:"owner"`                                     // Owner of the email
	ExpiryDate   time.Time `json:"expiry_date,omitempty" gorethink:"expiry_date,omitempty"`

	Type     string   `json:"type" gorethink:"type"` // auth/activate/code
	Scope    []string `json:"scope,omitempty" gorethink:"scope,omitempty"`
	ClientID string   `json:"client_id,omitempty" gorethink:"client_id,omitempty"`
}

func (t *Token) IsExpired() bool {
	return !t.ExpiryDate.IsZero() && t.ExpiryDate.Before(time.Now())
}
