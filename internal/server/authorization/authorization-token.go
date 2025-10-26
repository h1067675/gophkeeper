// Package authorization реализует функции авторизации пользователя по сессиям и токенам, а также функцию двухфакторной авторизации
package authorization

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

// GetSessionFromToken получает данные сессии из токена пользователя
func (a *Autorization) GetSessionFromToken(tokenString string) (string, error) {
	var cl = Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &cl, func(t *jwt.Token) (interface{}, error) {
		return []byte(a.SecretKey), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		err := fmt.Errorf("token is not valid")
		a.Logger.WithError(err).Info("token is not valid")
		return "", err
	}
	a.Logger.Infof("user id=%v restore from token", cl.Session)
	return cl.Session, nil
}

// CreateTokenSession создает токен с данными сессии
func (a *Autorization) CreateTokenSession(session string) (string, error) {
	var cl = Claims{
		Session: session,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)
	tokenString, err := token.SignedString([]byte(a.SecretKey))
	if err != nil {
		a.Logger.Debug("error token generate")
		return "", err
	}
	a.Logger.Debugf("create new token for user session=%v", cl.Session)
	return tokenString, nil
}
