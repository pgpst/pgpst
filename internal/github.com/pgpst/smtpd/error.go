package smtpd

import (
	"fmt"
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}
