package storage

import (
	"context"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
)

type RequestRegData struct {
	Login string `json:"login"`
	Pwd   string `json:"password"`
}

type Repositories interface {
	Close() error
	AddUser(ctx context.Context, regData RequestRegData) error
}

func NewStorage(cfg config.Flags) (Repositories, error) {
	db, err := NewDBStorage(cfg.DBURI)
	if err != nil {
		return nil, err
	}
	return db, nil
}
