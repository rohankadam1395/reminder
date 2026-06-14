package apperrors

import "errors"

var (
	ErrNotFound     = errors.New("reminder not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal server error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
)
