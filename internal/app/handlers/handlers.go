package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/accrual"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/authorizer"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/config"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/middleware"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/models"
	"github.com/Julia-ivv/loyalty-system.git/internal/app/storage"
	"github.com/go-chi/chi"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handlers struct {
	stor        storage.Repositories
	cfg         config.Flags
	accrualSyst accrual.AccrualSystem
}

func NewHandlers(stor storage.Repositories, cfg config.Flags, accSyst accrual.AccrualSystem) *Handlers {
	h := &Handlers{}
	h.stor = stor
	h.cfg = cfg
	h.accrualSyst = accSyst
	return h
}

func NewURLRouter(repo storage.Repositories, cfg config.Flags, accSyst accrual.AccrualSystem) chi.Router {
	hs := NewHandlers(repo, cfg, accSyst)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(hs.userRegistration)))
		r.Post("/api/user/login", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(hs.userAuthentication)))
		r.Post("/api/user/orders", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.postUserOrder))))
		r.Get("/api/user/orders", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.getUserOrders))))
		r.Get("/api/user/balance", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.getUserBalance))))
		r.Post("/api/user/balance/withdraw", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.postWithdraw))))
		r.Get("/api/user/withdrawals", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.GetUserWithdrawals))))
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

	var reqRegData models.RequestRegData
	err = json.Unmarshal(reqBody, &reqRegData)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.stor.RegUser(req.Context(), reqRegData)
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

	err = h.stor.AuthUser(req.Context(), models.RequestAuthData(reqRegData))
	if err != nil {
		var authErr *authorizer.AuthErr
		if (errors.As(err, &authErr)) && (authErr.ErrType == authorizer.InvalidHash) {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := authorizer.BuildToken(reqRegData.Login, reqRegData.Pwd)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:     authorizer.AccessToken,
		Value:    tokenString,
		Expires:  time.Now().Add(authorizer.TokenExp),
		Path:     "/",
		HttpOnly: true,
	})

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

	var reqAuthData models.RequestAuthData
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
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:    authorizer.AccessToken,
		Value:   tokenString,
		Expires: time.Now().Add(authorizer.TokenExp),
	})
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) postUserOrder(res http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(reqBody) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}

	orderNum := string(reqBody)
	isValidNum, err := LuhnCheck(orderNum)
	if (err != nil) || !isValidNum {
		http.Error(res, "incorrect order number format", http.StatusUnprocessableEntity)
		return
	}

	value := req.Context().Value(authorizer.UserContextKey)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	userLogin := value.(string)

	err = h.stor.PostUserOrder(req.Context(), orderNum, userLogin)
	if err != nil {
		var postErr *storage.StorErr
		if errors.As(err, &postErr) && postErr.ErrType == storage.UploadByAnotherUser {
			http.Error(res, err.Error(), http.StatusConflict)
			return
		}
		if errors.As(err, &postErr) && postErr.ErrType == storage.UploadByThisUser {
			res.WriteHeader(http.StatusOK)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	go h.accrualSyst.AddOrderForWork(orderNum)

	res.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) getUserOrders(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	userLogin := value.(string)

	orders, err := h.stor.GetUserOrders(req.Context(), userLogin)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(res, "204 No Content", http.StatusNoContent)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	resp, err := json.Marshal(orders)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) getUserBalance(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	userLogin := value.(string)

	balance, err := h.stor.GetUserBalance(req.Context(), userLogin)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	resp, err := json.Marshal(balance)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) postWithdraw(res http.ResponseWriter, req *http.Request) {
	reqJSON, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if len(reqJSON) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}

	var reqData models.RequestWithdrawData
	err = json.Unmarshal(reqJSON, &reqData)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	isValidNum, err := LuhnCheck(reqData.OrderNumber)
	if (err != nil) || !isValidNum {
		http.Error(res, "incorrect order number format", http.StatusUnprocessableEntity)
		return
	}

	value := req.Context().Value(authorizer.UserContextKey)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	userLogin := value.(string)

	err = h.stor.PostWithdraw(req.Context(), userLogin, reqData)
	if err != nil {
		var storErr *storage.StorErr
		if errors.As(err, &storErr) && storErr.ErrType == storage.NotEnoughPoints {
			http.Error(res, err.Error(), http.StatusPaymentRequired)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetUserWithdrawals(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	userLogin := value.(string)

	resData, err := h.stor.GetUserWithdrawals(req.Context(), userLogin)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resData) == 0 {
		http.Error(res, "no withdrawals", http.StatusNoContent)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	resp, err := json.Marshal(resData)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
