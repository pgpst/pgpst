package models

import (
	"time"
)

type Thread struct {
	ID           string    `json:"id" gorethink:"id"`                                           // 20-char id
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // time of creation
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // time of last mod
	Owner        string    `json:"owner" gorethink:"owner"`                                     // Owner of the email

	Emails  []string `json:"emails" gorethink:"emails"`
	Labels  []string `json:"labels" gorethink:"labels"`
	Members []string `json:"members" gorethink:"members"`

	IsRead   string `json:"is_read" gorethink:"is_read"`
	LastRead string `json:"last_read" gorethink:"last_read"`

	IsSecure string `json:"is_secure" gorethink:"is_secure"`

	Manifest []byte `json:"manifest,omitempty" gorethink:"manifest,omitempty"`
}
