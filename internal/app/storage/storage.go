package storage

import (
	"context"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
)

type Customer interface {
	RegUser(ctx context.Context, regData RequestRegData) error
	AuthUser(ctx context.Context, authData RequestAuthData) error
}

type OrdersWorker interface {
	PostUserOrder(ctx context.Context, orderNumber string, userLogin string) error
	GetUserOrders(ctx context.Context, userLogin string) ([]ResponseOrder, error)
}

type PointsWorker interface {
	GetUserBalance(ctx context.Context, userLogin string) (ResponseBalance, error)
	PostWithdraw(ctx context.Context, userLogin string, withdrawData RequestWithdrawData) error
	GetUserWithdrawals(ctx context.Context, userLogin string) ([]ResponseWithdrawals, error)
	UpdateUserAccrual(ctx context.Context, newData ResponseAccrual) error
}

type Repositorier interface {
	Close() error
	Customer
	OrdersWorker
	PointsWorker
}

func NewStorage(cfg config.Flags) (Repositorier, error) {
	db, err := NewDBStorage(cfg.DBURI)
	if err != nil {
		return nil, err
	}
	return db, nil
}
