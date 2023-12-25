package authorizer

import "fmt"

type TypeAuthErrors string

const (
	QeuryError  TypeAuthErrors = "request completed with error"
	InvalidHash TypeAuthErrors = "invalid hash"
)

type AuthErr struct {
	ErrType TypeAuthErrors
	Err     error
}

func (e *AuthErr) Error() string {
	return fmt.Sprintln(e.ErrType)
}

func NewAuthError(t TypeAuthErrors, err error) error {
	return &AuthErr{
		ErrType: t,
		Err:     err,
	}
}
