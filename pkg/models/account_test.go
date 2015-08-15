package models_test

import (
	"errors"
	"testing"

	"github.com/pgpst/pgpst/internal/github.com/pzduniak/mcf"
	. "github.com/pgpst/pgpst/internal/github.com/smartystreets/goconvey/convey"

	"github.com/pgpst/pgpst/pkg/models"
)

type dumbEncoder struct {
	CreateByteR     []byte
	CreateErrorR    error
	VerifyBoolR     bool
	VerifyErrorR    error
	IsCurrentBoolR  bool
	IsCurrentErrorR error
}

func (d *dumbEncoder) Id() []byte {
	return []byte("dumb")
}
func (d *dumbEncoder) Create(plain []byte) ([]byte, error) {
	return d.CreateByteR, d.CreateErrorR
}
func (d *dumbEncoder) Verify(plain []byte, encoded []byte) (bool, error) {
	return d.VerifyBoolR, d.VerifyErrorR
}
func (d *dumbEncoder) IsCurrent(encoded []byte) (bool, error) {
	return d.IsCurrentBoolR, d.IsCurrentErrorR
}

func TestAccount(t *testing.T) {
	Convey("Given an account and a dumb encoder", t, func() {
		account := &models.Account{
			Password: []byte("hi"),
		}

		encoder := &dumbEncoder{}

		Convey("Basic checks should succeed", func() {
			// We can't split into multiple conveys, because I don't think mcf supports multithreading
			// err == nil
			mcf.SetDefault(mcf.SCRYPT)
			So(account.SetPassword([]byte{0}), ShouldBeNil)

			// err == nil
			valid, updated, err := account.VerifyPassword([]byte{0})
			So(valid, ShouldBeTrue)
			So(updated, ShouldBeFalse)
			So(err, ShouldBeNil)

			// valid == false
			valid, updated, err = account.VerifyPassword([]byte{1})
			So(valid, ShouldBeFalse)
			So(updated, ShouldBeFalse)
			So(err, ShouldBeNil)

			// back to dumb encoder, err != nil
			mcf.Register(mcf.PBKDF2, encoder)
			mcf.SetDefault(mcf.PBKDF2)

			// err != nil
			encoder.CreateByteR = nil
			encoder.CreateErrorR = errors.New("hello1")
			So(account.SetPassword([]byte("sth")), ShouldNotBeNil)

			// set the password properly
			encoder.CreateByteR = []byte("$dumbsth")
			encoder.CreateErrorR = nil
			So(account.SetPassword([]byte("dumb0")), ShouldBeNil)

			// err is not nil in verify
			encoder.VerifyBoolR = false
			encoder.VerifyErrorR = errors.New("hello")
			valid, updated, err = account.VerifyPassword([]byte("dumb0"))
			So(valid, ShouldBeFalse)
			So(updated, ShouldBeFalse)
			So(err, ShouldNotBeNil)

			// verify and update no err
			encoder.VerifyBoolR = true
			encoder.VerifyErrorR = nil
			encoder.IsCurrentBoolR = false
			encoder.IsCurrentErrorR = nil
			valid, updated, err = account.VerifyPassword([]byte("dumb0"))
			So(valid, ShouldBeTrue)
			So(updated, ShouldBeTrue)
			So(err, ShouldBeNil)

			// verify and updat err
			encoder.CreateErrorR = errors.New("hello")
			valid, updated, err = account.VerifyPassword([]byte("dumb0"))
			So(valid, ShouldBeTrue)
			So(updated, ShouldBeFalse)
			So(err, ShouldNotBeNil)

			// failing iscurrent
			encoder.IsCurrentBoolR = false
			encoder.IsCurrentErrorR = errors.New("hello2")
			valid, updated, err = account.VerifyPassword([]byte{0})
			So(valid, ShouldBeFalse)
			So(updated, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

	})
}
