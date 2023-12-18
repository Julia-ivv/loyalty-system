package storage

import (
	"context"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/models"
)

type Repositories interface {
	Close() error
	RegUser(ctx context.Context, regData models.RequestRegData) error
	AuthUser(ctx context.Context, authData models.RequestAuthData) error
	PostUserOrder(ctx context.Context, orderNumber string, userLogin string) error
	GetUserOrders(ctx context.Context, userLogin string) ([]models.ResponseOrder, error)
	GetUserBalance(ctx context.Context, userLogin string) (models.ResponseBalance, error)
	PostWithdraw(ctx context.Context, userLogin string, withdrawData models.RequestWithdrawData) error
	GetUserWithdrawals(ctx context.Context, userLogin string) ([]models.ResponseWithdrawals, error)
	UpdateUserAccrual(ctx context.Context, newData models.ResponseAccrual) error
}

func NewStorage(cfg config.Flags) (Repositories, error) {
	db, err := NewDBStorage(cfg.DBURI)
	if err != nil {
		return nil, err
	}
	return db, nil
}
