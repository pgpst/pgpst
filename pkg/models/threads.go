package models

import (
	"time"
)

type Thread struct {
	ID           string    `db:"id"`
	DateCreated  time.Time `db:"date_created"`
	DateModified time.Time `db:"date_modified"`
	Owner        string    `db:"owner"`
	Address      string    `db:"address"`
	SMembers     string    `db:"members"`
	Members      [][]byte  `db:"-"`
	IsRead       bool      `db:"is_read"`
	LastRead     string    `db:"last_read"`
	Secure       string    `db:"secure"`
}

type ThreadLabel struct {
	ThreadID string `db:"thread_id"`
	LabelID  string `db:"label_id"`
}
