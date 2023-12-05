package storage

import (
	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
)

type Repositories interface {
	Close() error
}

func NewStorage(cfg config.Flags) (Repositories, error) {
	db, err := NewDBStorage(cfg.DBURI)
	if err != nil {
		return nil, err
	}
	return db, nil
}
