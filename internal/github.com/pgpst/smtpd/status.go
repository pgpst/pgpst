package smtpd

var statusString = map[int]string{
	StatusPasswordNeeded:     "4.7.12  A password transition is needed",
	StatusMessageError:       "Requested mail action not taken",
	StatusTempAuthFailure:    "4.7.0  Temporary authentication failure",
	StatusAuthInvalid:        "5.7.8  Authentication credentials invalid",
	StatusAuthRequired:       "5.7.0  Authentication required",
	StatusEncryptionRequired: "5.7.11  Encryption required for requested authentication mechanism",
	StatusServerError:        "Requested mail action not taken: server error",
	StatusExceedStorage:      "Requested mail action aborted: exceeded storage allocation",
}

func StatusString(status int) string {
	s, found := statusString[status]
	if !found {
		return "unknown"
	}
	return s
}

const (
	StatusPasswordNeeded       = 432
	StatusMessageError         = 450
	StatusMessageExceedStorage = 452
	StatusTempAuthFailure      = 454
	StatusAuthInvalid          = 535
	StatusAuthRequired         = 530
	StatusEncryptionRequired   = 538
	StatusServerError          = 550
	StatusExceedStorage        = 552
)

var (
	ErrPasswordNeeded       = Error{Code: StatusPasswordNeeded, Message: StatusString(StatusPasswordNeeded)}
	ErrMessageError         = Error{Code: StatusMessageError, Message: StatusString(StatusMessageError)}
	ErrMessageExceedStorage = Error{Code: StatusMessageExceedStorage, Message: StatusString(StatusMessageExceedStorage)}
	ErrTempAuthFailure      = Error{Code: StatusTempAuthFailure, Message: StatusString(StatusTempAuthFailure)}
	ErrAuthInvalid          = Error{Code: StatusAuthInvalid, Message: StatusString(StatusAuthInvalid)}
	ErrAuthRequired         = Error{Code: StatusAuthRequired, Message: StatusString(StatusAuthRequired)}
	ErrServerError          = Error{Code: StatusServerError, Message: StatusString(StatusServerError)}
	ErrExceedStorage        = Error{Code: StatusExceedStorage, Message: StatusString(StatusExceedStorage)}
)
