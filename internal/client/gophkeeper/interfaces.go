// Package gophkeeper реализует основную бизнес логику приложения
// описывает интерфейсы для доступа к микросервисам
package gophkeeper

import (
	"gophkeeper/pkg/domain"
	"time"
)

type (

	// Configurer интерфейсный тип описывает сервис конфигурирования
	Configurer interface {
		GetServerAddress() (serverAddress string, err error)
		GetGRPCAddress() (address string)
		GetHTTPSAddress() (address string)
		CheckNetAddress(str string) error
		SetConfig(serverAddress, grpcPort, httpsPort string) error
		CheckConfig() error
		SaveUserSession(session string) error
		GetUserSession() (token string)
		DeleteUserSession() error
		SaveUserSaltB64(session string) error
		GetUserSaltB64() (token string)
		DeleteUserSaltB64() error
		SaveUserSPassHash(session string) error
		GetUserSPassHash() (token string)
		DeleteUserSPassHash() error
	}

	// Repository интерфейсный тип описывает сервис хранения данных
	Repository interface {
		DropMigrations() error
		GetLastSync() (time.Time, error)
		UpdateLastSync(time.Time) error

		NewPassword(card *domain.Password) error
		GetPassword(id int) (domain.Password, error)
		GetPasswordsList(from time.Time) ([]domain.Password, error)
		SavePassword(pass domain.Password) error
		SyncPassword(pass domain.Password) error
		DeletePassword(pass domain.Password) error

		NewCard(card *domain.Card) error
		GetCard(id int) (domain.Card, error)
		GetCardsList(from time.Time) ([]domain.Card, error)
		SaveCard(card domain.Card) error
		SyncCard(card domain.Card) error
		DeleteCard(card domain.Card) error

		NewTextData(txt *domain.Text) error
		GetTextData(id int) (domain.Text, error)
		GetTextDataList(from time.Time) ([]domain.Text, error)
		SaveTextData(txt domain.Text) error
		SyncText(txt domain.Text) error
		DeleteTextData(txt domain.Text) error

		NewBinaryData(bin *domain.Binary) error
		GetBinaryData(id int) (domain.Binary, error)
		GetBinaryDataList(from time.Time) ([]domain.Binary, error)
		SaveBinaryData(bin domain.Binary) error
		SyncBinary(bin domain.Binary) error
		DeleteBinaryData(bin domain.Binary) error
	}

	// Crypter интерфейсный тип описывает сервис шифрования
	Crypter interface {
		Argon2id(data, salt []byte, b64Salt string) (string, error)
		CompareHash(data string, encodedHash string) (ok bool, err error)
		Encrypt(data []byte, password, salt64 string) (nonceAndCipherData string, err error)
		Decrypt(nonceAndCipherData, password, salt64 string) (data []byte, err error)
	}

	// Transporter интерфейсный тип описывает сервис передачи данных
	Transporter interface {
		Registration(user *domain.User) (textErr string, err error)
		Authorization(user *domain.User) (textErr string, err error)
		TOTPConfirmation(user *domain.User, code string) (textErr string, err error)
		TOTPUpdate(user *domain.User, code string) (textErr string, err error)
		SetServers(servergrpc, serverhttps string) error
		SaveHashSecretPassword(user domain.User, hash string) (textErr string, err error)
		SyncData(dataOut domain.SyncPayload) (dataIn domain.SyncPayload, textErr string, err error)
	}

	// QRGererator интерфейсный тип описывает сервис формирования QR кодов
	QRGererator interface {
		QRMediumToString(qrString string) (string, error)
	}
)
