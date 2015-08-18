package api

const (
	CodeGeneralUnknown = 1000 + iota
	CodeGeneralUnimplemented
	CodeGeneralInvalidInput
	CodeGeneralDatabaseError
	CodeGeneralInvalidAction
)

const (
	CodeOAuthUnknown = 2000 + iota
	CodeOAuthInvalidApplication
	CodeOAuthInvalidSecret
	CodeOAuthInvalidCode
	CodeOAuthValidationFailed
	CodeOAuthInvalidAddress
	CodeOAuthInvalidPassword
)
