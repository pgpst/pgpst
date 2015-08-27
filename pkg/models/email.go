package models

import (
	"time"
)

type Email struct {
	ID           string    `json:"id" gorethink:"id"`                                           // 20-char id
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // time of creation
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // time of last mod
	Owner        string    `json:"owner" gorethink:"owner"`                                     // Owner of the email

	MessageID string `json:"message_id" gorethink:"message_id"`
	/*From string   `json:"from" gorethink:"from"`                   // who sent it
	To   []string `json:"to" gorethink:"to"`                       // who's the recipient
	CC   []string `json:"cc,omitempty" gorethink:"cc,omitempty"`   // carbon copy
	BCC  []string `json:"bcc,omitempty" gorethink:"bcc,omitempty"` // blind carbon copy*/

	Thread string `json:"thread" gorethink:"thread"` // thread id
	Status string `json:"status" gorethink:"status"` // status - received, sent or sending

	Manifest []byte `json:"manifest" gorethink:"manifest"` // Description of the body including keys
	Body     []byte `json:"body" gorethink:"body"`         // Email's body encrypted using AES256-CTR
}
