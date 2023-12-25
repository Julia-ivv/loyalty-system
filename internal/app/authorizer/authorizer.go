package authorizer

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const SecretKey = "byrhtvtyn"
const AccessToken = "accessToken"
const TokenExp = time.Hour * 3

type key string

const UserContextKey key = "user"

type Claims struct {
	jwt.RegisteredClaims
	Login string
	Pwd   string
}

func BuildToken(userLogin, userPwd string) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			},
			Login: userLogin,
			Pwd:   userPwd,
		})
	tokenString, err = token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GetUserDataFromToken(tokenString string) (userLogin string, userPwd string, err error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if (err != nil) || (!token.Valid) {
		return "", "", err
	}
	return claims.Login, claims.Pwd, nil
}
