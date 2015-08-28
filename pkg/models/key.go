package models

import (
	"time"
)

type Key struct {
	ID           string    `json:"id" gorethink:"id"`                                           // key's fingerprint
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // when it was created
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // last update
	Owner        string    `json:"owner,omitempty" gorethink:"owner,omitempty"`                 // owner of the key

	Algorithm        uint8  `json:"algorithm,omitempty" gorethink:"algorithm,omitempty"` // algorithm of the key
	Length           uint16 `json:"length,omitempty" gorethink:"length,omitempty"`       // key's length
	Body             []byte `json:"body,omitempty" gorethink:"body,omitempty"`           // the actual key
	KeyID            uint64 `json:"key_id,omitempty" gorethink:"key_id,omitempty"`
	KeyIDString      string `json:"key_id_string,omitempty" gorethink:"key_id_string,omitempty"`             // key_id
	KeyIDShortString string `json:"key_id_short_string,omitempty" gorethink:"key_id_short_string,omitempty"` // shorter version of key_id
	MasterKey        string `json:"master_key,omitempty" gorethink:"master_key,omitempty"`                   // master key

	Identities []*Identity `json:"identities,omitempty" gorethink:"identities,omitempty"`
}

type Identity struct {
	Name          string       `json:"name" gorethink:"name"`
	SelfSignature *Signature   `json:"self_signature" gorethink:"self_signature"`
	Signatures    []*Signature `json:"signatures" gorethink:"signatures"`
}

type Signature struct {
	Type         uint8     `json:"type" gorethink:"type"`
	Algorithm    uint8     `json:"algorithm" gorethink:"algorithm"`
	Hash         uint      `json:"hash" gorethink:"hash"`
	CreationTime time.Time `json:"creation_time" gorethink:"creation_time"`

	SigLifetimeSecs uint32 `json:"sig_lifetime_secs,omitempty" gorethink:"sig_lifetime_secs,omitempty"`
	KeyLifetimeSecs uint32 `json:"key_lifetime_secs,omitempty" gorethink:"key_lifetime_secs,omitempty"`
	IssuerKeyID     uint64 `json:"issuer_key_id,omitempty" gorethink:"issuer_key_id,omitempty"`
	IsPrimaryID     bool   `json:"is_primary_id,omitempty" gorethink:"is_primary_id,omitempty"`

	RevocationReason     uint8  `json:"revocation_reason,omitempty" gorethink:"revocation_reason,omitempty"`
	RevocationReasonText string `json:"revocation_reason_text,omitempty" gorethink:"revocation_reason_text,omitempty"`
}
