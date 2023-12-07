package middleware

import (
	"net/http"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/authorizer"
)

func HandlerWithAuth(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			token, err := req.Cookie(authorizer.AccessToken)
			if err != nil {
				http.Error(res, "401 Unauthorized", http.StatusUnauthorized)
				return
			}

			_, _, err = authorizer.GetUserDataFromToken(token.Value)
			if err != nil {
				http.Error(res, "401 Unauthorized", http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(res, req)
		})
}
