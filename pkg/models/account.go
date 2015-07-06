package models

import (
	"time"

	"github.com/pgpst/pgpst/internal/github.com/gyepisam/mcf"
	_ "github.com/pgpst/pgpst/internal/github.com/gyepisam/mcf/scrypt"
)

type Account struct {
	ID           string    `json:"id" gorethink:"id"`                                           // 20-char long id
	DateCreated  time.Time `json:"date_created,omitempty" gorethink:"date_created,omitempty"`   // when the account was created
	DateModified time.Time `json:"date_modified,omitempty" gorethink:"date_modified,omitempty"` // last modification
	MainAddress  string    `json:"main_address" gorethink:"main_address"`                       // main address id
	Password     string    `json:"-" gorethink:"password,omitempty"`                            // scrypt'd sha256 password
	Subscription string    `json:"subscription" gorethink:"subscription"`                       // chosen subscription
	AltEmail     string    `json:"alt_email" gorethink:"alt_email"`                             // alternative email
	Status       string    `json:"status" gorethink:"status"`                                   // account's status
}

func (a *Account) VerifyPassword(password string) (bool, bool, error) {
	valid, err := mcf.Verify(password, a.Password)
	if err != nil {
		return false, false, err
	}
	if !valid {
		return false, false, nil
	}

	current, err := mcf.IsCurrent(a.Password)
	if err != nil {
		return false, false, err
	}

	if !current {
		err := a.SetPassword(password)
		if err != nil {
			return true, false, err
		}

		a.DateModified = time.Now()

		return true, true, nil
	}

	return true, false, nil
}

func (a *Account) SetPassword(password string) error {
	encrypted, err := mcf.Create(password)
	if err != nil {
		return err
	}

	a.Password = encrypted
	a.DateModified = time.Now()

	return nil
}
