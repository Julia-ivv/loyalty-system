package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/authorizer"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/storage"
	"github.com/go-chi/chi"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	hs := NewHandlers(repo, cfg)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", hs.userRegistration)
		r.Post("/api/user/login", hs.userAuthentication)
	})

	return r
}

func (h *Handlers) userRegistration(res http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(reqBody) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}

	var reqRegData storage.RequestRegData
	err = json.Unmarshal(reqBody, &reqRegData)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.stor.AddUser(req.Context(), reqRegData)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			http.Error(res, err.Error(), http.StatusConflict)
			return
		} else {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) userAuthentication(res http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(reqBody) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}

	var reqAuthData storage.RequestAuthData
	err = json.Unmarshal(reqBody, &reqAuthData)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.stor.AuthUser(req.Context(), reqAuthData)
	if err != nil {
		var authErr *authorizer.AuthErr
		if (errors.As(err, &authErr)) && (authErr.ErrType == authorizer.InvalidHash) {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := authorizer.BuildToken(reqAuthData.Login, reqAuthData.Pwd)
	http.SetCookie(res, &http.Cookie{
		Name:    authorizer.AccessToken,
		Value:   tokenString,
		Expires: time.Now().Add(authorizer.TokenExp),
	})
	res.WriteHeader(http.StatusOK)
}
