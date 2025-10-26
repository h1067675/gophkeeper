// Package gophkeeper реализует основную бизнес логику сервера
// описывает интерфейсы для доступа к микросервисам
package gophkeeper

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// структуры
type (
	// GophKeeperService описывает структуру зависимостей для доступа к базе данных, конфигурации и серверам
	Service struct {
		Repository  Repository
		Authorizer  Authorizer
		Crypter     Crypter
		Mailer      Mailer
		CompanyName string
		Logger      *log.Logger
	}
)

// NewService инициализирует новый сервис бизнес логики
func NewService(companyName string, depo Repository, auth Authorizer, crypt Crypter, mail Mailer, logger *log.Logger) (*Service, error) {
	var app = &Service{
		Repository:  depo,
		Authorizer:  auth,
		Crypter:     crypt,
		Mailer:      mail,
		Logger:      logger,
		CompanyName: companyName,
	}
	if app.Repository == nil {
		return nil, errors.New("failed to create application, depository is missing")
	}
	if app.Authorizer == nil {
		return nil, errors.New("failed to create application, authorization module is missing")
	}

	return app, nil
}
