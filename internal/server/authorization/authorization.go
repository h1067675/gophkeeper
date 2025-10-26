// Package authorization реализует функции авторизации пользователя по сессиям и токенам, а также функцию двухфакторной авторизации
package authorization

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// структуры
type (

	// Autorization описывает структуру модуля авторизации
	Autorization struct {
		SecretKey string
		Logger    *log.Logger
	}

	// Claims описывает структуру для создания токена из кода сессии и дальнейшей расшифровки
	Claims struct {
		jwt.RegisteredClaims
		Session string
	}
)

// Init инициализирует модуль авторизации
func Init(secretKey string, logger *log.Logger) (*Autorization, error) {
	if secretKey == "" {
		return nil, errors.New("empty secret key")
	}
	if len([]byte(secretKey)) < 10 {
		return nil, errors.New("length secret key is less than 10 characters")
	}
	return &Autorization{SecretKey: secretKey, Logger: logger}, nil
}
