package models

import (
	"time"
)

type Resource struct {
	ID           string    `json:"id" gorethink:"id"`                                           // 20-char id
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // time of creation
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // time of last mod
	Owner        string    `json:"owner" gorethink:"owner"`                                     // Owner of the email

	Meta map[string]interface{} `json:"meta,omitempty" gorethink:"meta,omitempty"`
	Body []byte                 `json:"body" gorethink:"body"`
	Tags []string               `json:"tags,omitempty" gorethink:"tags,omitempty"`
}
