package v1

import "errors"

var (
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrInvalidAccessKey = errors.New("invalid access key")
	ErrEmailMismatch    = errors.New("access key does not match email")
	ErrAlreadyRevoked   = errors.New("access key already revoked")
)
