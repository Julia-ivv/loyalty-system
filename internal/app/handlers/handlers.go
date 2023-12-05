package handlers

import (
	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/storage"
	"github.com/go-chi/chi"
)

type Handlers struct {
	stor storage.Repositories
	cfg  config.Flags
}

func NewHandlers(stor storage.Repositories, cfg config.Flags) *Handlers {
	h := &Handlers{}
	h.stor = stor
	h.cfg = cfg
	return h
}

func NewURLRouter(repo storage.Repositories, cfg config.Flags) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {

	})

	return r
}
