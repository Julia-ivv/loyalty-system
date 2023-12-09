package middleware

import (
	"context"
	"net/http"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/authorizer"
)

func HandlerWithAuth(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			token, err := req.Cookie(authorizer.AccessToken)
			if err != nil {
				http.Error(res, err.Error(), http.StatusUnauthorized)
				return
			}

			login, _, err := authorizer.GetUserDataFromToken(token.Value)
			if err != nil {
				http.Error(res, err.Error(), http.StatusUnauthorized)
				return
			}
			newctx := context.WithValue(req.Context(), authorizer.UserContextKey, login)

			h.ServeHTTP(res, req.WithContext(newctx))
		})
}
