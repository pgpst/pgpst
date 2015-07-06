package models

import (
	"time"
)

type Email struct {
	ID           string    `json:"id" gorethink:"id"`                                           // 20-char id
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // time of creation
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // time of last mod
	Owner        string    `json:"owner" gorethink:"owner"`                                     // Owner of the email

	From string   `json:"from" gorethink:"from"`                   // who sent it
	To   []string `json:"to" gorethink:"to"`                       // who's the recipient
	CC   []string `json:"cc,omitempty" gorethink:"cc,omitempty"`   // carbon copy
	BCC  []string `json:"bcc,omitempty" gorethink:"bcc,omitempty"` // blind carbon copy

	Manifest []byte   `json:"manifest,omitempty" gorethink:"manifest,omitempty"` // manifest body
	Files    []string `json:"files,omitempty" gorethink:"files,omitempty"`       // resource ids
	Body     []byte   `json:"body,omitempty" gorethink:"body,omitempty"`         // main body parts

	Thread string `json:"thread" gorethink:"thread"` // thread id
	Status string `json:"status" gorethink:"status"` // status - received, sent or sending
	Secure bool   `json:"secure" gorethink:"secure"` // encrypted true/false
}
