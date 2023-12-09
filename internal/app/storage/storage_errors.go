package storage

import "fmt"

type TypeStorErrors string

const (
	UploadByThisUser    TypeStorErrors = "the order has already been uploaded by this user"
	UploadByAnotherUser TypeStorErrors = "the order has already been uploaded by another user"
)

type StorErr struct {
	ErrType TypeStorErrors
	Err     error
}

func (e *StorErr) Error() string {
	return fmt.Sprintln(e.ErrType)
}

func NewStorError(t TypeStorErrors, err error) error {
	return &StorErr{
		ErrType: t,
		Err:     err,
	}
}
