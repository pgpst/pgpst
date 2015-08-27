package smtpd_test

import (
	"testing"

	"github.com/pgpst/pgpst/internal/github.com/pgpst/smtpd"
)

func TestStatus(t *testing.T) {
	if smtpd.ErrServerError.Error() != "550 Requested mail action not taken: server error" {
		t.Fatal("Invalid status message")
	}
}
