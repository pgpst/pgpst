package models

import (
	"time"
)

type PublicKey struct {
	ID           string         `db:"id"`
	Fingerprint  string         `db:"fingerprint"`
	IDShort      string         `db:"id_short"`
	DateCreated  time.Time      `db:"date_created"`
	DateModified time.Time      `db:"date_modified"`
	Owner        string         `db:"owner"`
	Body         string         `db:"body"`
	Algorithm    int            `db:"algorithm"`
	Length       int            `db:"length"`
	SIdentities  string         `db:"identities"`
	Identities   []*KeyIdentity `db:"-"`
}

type KeyIdentity struct {
	Name          string          `json:"name"`
	SelfSignature *KeySignature   `json:"self_signature"`
	Signatures    []*KeySignature `json:"signatures"`
}

type KeySignature struct {
	CreationTime time.Time `json:"creation_time"`
	ExpiryTime   time.Time `json:"expiry_time"`
	Issuer       string    `json:"issuer"`
	Algorithm    int       `json:"algorithm"`
	Hash         int       `json:"hash"`
	Type         int       `json:"type"`
}
