package models

import (
	"net/mail"
)

type EmailNode struct {
	Headers        mail.Header  `json:"headers"`
	BasePosition   int          `json:"base_position"`
	HeaderPosition [2]int       `json:"header_position"`
	BodyPosition   [2]int       `json:"body_position"`
	Children       []*EmailNode `json:"children,omitempty"`
}

type Manifest struct {
	Key         []byte     `json:"key"`
	Nonce       []byte     `json:"nonce"`
	Description *EmailNode `json:"description"`
}
