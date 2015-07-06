package models

import (
	"time"
)

type Label struct {
	ID           string    `json:"id" gorethink:"id"`                                           // 20-char id
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // time of creation
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // time of last mod
	Owner        string    `json:"owner" gorethink:"owner"`                                     // Owner of the email

	Name   string `json:"name" gorethink:"name"`
	System bool   `json:"is_system" gorethink:"is_system"`

	UnreadThreads int `json:"unread_threads" gorethink:"unread_threads,omitempty"`
	TotalThreads  int `json:"total_threads" gorethink:"total_threads,omitempty"`
}
