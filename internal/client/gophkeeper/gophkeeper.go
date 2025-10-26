// Package gophkeeper реализует основную бизнес логику приложения
// описывает интерфейсы для доступа к микросервисам
package gophkeeper

import (
	"errors"
	"gophkeeper/pkg/domain"

	log "github.com/sirupsen/logrus"
)

// структуры
type (

	// Service описывает структуру зависимостей для доступа к базе данных, конфигурации, шифрования и транспортным сервисам
	Service struct {
		Configurer  Configurer
		Repository  Repository
		Transporter Transporter
		Crypter     Crypter
		QRGenerator QRGererator

		Logger *log.Logger

		User     domain.User
		Versions Versions
	}

	// Versions описывает формат хранения информации о версии для отображения пользователю
	Versions struct {
		BuildVersion, BuildDate, BuildCommit string
	}
)

// NewService инициализирует сервис бизнес логики
func NewService(versions Versions, conf Configurer, depo Repository, crypt Crypter, client Transporter, qrgen QRGererator, logger *log.Logger) (*Service, error) {
	var app = &Service{
		Configurer:  conf,
		Repository:  depo,
		Transporter: client,
		Crypter:     crypt,
		QRGenerator: qrgen,
		Logger:      logger,
		Versions:    versions,
	}
	app.User = domain.User{Session: app.GetSession(), SaltB64: app.GetSaltB64(), SPassHash: app.GetSPassHas()}

	if app.Repository == nil {
		return nil, errors.New("failed to create application, depository is missing")
	}

	return app, nil
}
