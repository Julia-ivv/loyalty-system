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

type Repositories interface {
	Close() error
	RegUser(ctx context.Context, regData RequestRegData) error
	AuthUser(ctx context.Context, authData RequestAuthData) error
	PostOrder(ctx context.Context, orderNumber string, userLogin string) error
	GetUserOrders(ctx context.Context, userLogin string) ([]ResponseOrder, error)
}

func NewStorage(cfg config.Flags) (Repositories, error) {
	db, err := NewDBStorage(cfg.DBURI)
	if err != nil {
		return nil, err
	}
	return db, nil
}
