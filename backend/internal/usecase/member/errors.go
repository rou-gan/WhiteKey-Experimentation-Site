package member

import "errors"

var (
	ErrConflict     = errors.New("conflict")
	ErrAuthFailed   = errors.New("authentication failed")
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)
