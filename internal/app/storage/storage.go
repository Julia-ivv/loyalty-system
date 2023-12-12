package storage

import (
	"context"
	"time"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
)

type RequestRegData struct {
	Login string `json:"login"`
	Pwd   string `json:"password"`
}

type RequestAuthData struct {
	Login string `json:"login"`
	Pwd   string `json:"password"`
}

type ResponseOrder struct {
	Number       string    `json:"number"`
	Status       string    `json:"status"`
	Accrual      int       `json:"accrual,omitempty"`
	UploadedTime time.Time `json:"uploaded_at"`
}

type ResponseBalance struct {
	PointsBalance int `json:"current"`
	PointsUsed    int `json:"withdrawn"`
}

type RequestWithdrawData struct {
	OrderNumber string `json:"order"`
	WithdrawSum int    `json:"sum"`
}

type ResponseWithdrawals struct {
	OrderNumber  string    `json:"order"`
	WithdrawSum  int       `json:"sum"`
	WithdrawTime time.Time `json:"processed_at"`
}

type Repositories interface {
	Close() error
	RegUser(ctx context.Context, regData RequestRegData) error
	AuthUser(ctx context.Context, authData RequestAuthData) error
	PostUserOrder(ctx context.Context, orderNumber string, userLogin string) error
	GetUserOrders(ctx context.Context, userLogin string) ([]ResponseOrder, error)
	GetUserBalance(ctx context.Context, userLogin string) (ResponseBalance, error)
	PostWithdraw(ctx context.Context, userLogin string, withdrawData RequestWithdrawData) error
	GetUserWithdrawals(ctx context.Context, userLogin string) ([]ResponseWithdrawals, error)
}

func NewStorage(cfg config.Flags) (Repositories, error) {
	db, err := NewDBStorage(cfg.DBURI)
	if err != nil {
		return nil, err
	}
	return db, nil
}
