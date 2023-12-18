package accrual

import "fmt"

type TypeAccrualErrors string

const (
	NotRegistered   TypeAccrualErrors = "the order is not registered in the payment system"
	InternalError   TypeAccrualErrors = "internal server error"
	TooManyRequests TypeAccrualErrors = "too many requests"
)

type AccrualErr struct {
	ErrType TypeAccrualErrors
	Err     error
}

func (e *AccrualErr) Error() string {
	return fmt.Sprintln(e.ErrType)
}

func NewAccrualError(t TypeAccrualErrors, err error) error {
	return &AccrualErr{
		ErrType: t,
		Err:     err,
	}
}
