package storage

import (
	"context"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/models"
)

type Customer interface {
	RegUser(ctx context.Context, regData models.RequestRegData) error
	AuthUser(ctx context.Context, authData models.RequestAuthData) error
}

type OrdersWorker interface {
	PostUserOrder(ctx context.Context, orderNumber string, userLogin string) error
	GetUserOrders(ctx context.Context, userLogin string) ([]models.ResponseOrder, error)
}

type PointsWorker interface {
	GetUserBalance(ctx context.Context, userLogin string) (models.ResponseBalance, error)
	PostWithdraw(ctx context.Context, userLogin string, withdrawData models.RequestWithdrawData) error
	GetUserWithdrawals(ctx context.Context, userLogin string) ([]models.ResponseWithdrawals, error)
	UpdateUserAccrual(ctx context.Context, newData models.ResponseAccrual) error
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
