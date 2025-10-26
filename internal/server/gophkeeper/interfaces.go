// Package gophkeeper реализует основную бизнес логику сервера
// описывает интерфейсы для доступа к микросервисам
package gophkeeper

import (
	"time"

	"gophkeeper/pkg/domain"

	"github.com/pquerna/otp"
)

type (

	// Repository интерфейсный тип описывает сервис хранения данных
	Repository interface {
		UserRegistration(user domain.User) (userID int, err error)
		UserGetByLogin(login string) (userID int, err error)
		UserGetByEmail(email string) (userID int, err error)
		UserAuthorization(login string) (userID int, password, saltb64, spassHash string, err error)
		UserGetByToken(token string) (userID int, expired time.Time, err error)
		SaveUserPasswordHash(allowedToken, hash string) error

		SessionSave(userid int, session string, expire time.Time, allowed bool) error
		SessionSaveTOTPSecret(session, secretKey string) error
		SessionGetTOTPSecret(session string) (userID int, secretKey string, err error)
		SessionGet(userID int) (session, secretKey string, expire time.Time, err error)
		SessionGetEmail(session string) (email, code string, expireCode time.Time, err error)
		SessionSetEmailCode(session, code string, expireCode time.Time) (err error)

		Upsert(obj any, table string, conflictCol string) (int, error)
		Select(dest any, query string, args ...any) error
	}

	// Authorizer интерфейсный тип описывает сервис авторизации
	Authorizer interface {
		GetSessionFromToken(tokenString string) (session string, err error)
		CreateTokenSession(session string) (token string, err error)
		GenerateTOTPSecret(issuer string, email string) (key *otp.Key, err error)
		ValidateTOTPCode(code string, secretKey string) error
		GenerateSession(allowed bool) (session string, expire time.Time)
	}

	// Crypter интерфейсный тип описывает сервис шифрования
	Crypter interface {
		HashArgon2id(data string) (string, error)
		CompareHash(data string, encodedHash string) (ok bool, err error)
		GenerateSalt() (salt []byte, saltb64 string, err error)
	}

	// Mailer интерфейсный тип описывает сервис отправки кодов подтверждения
	Mailer interface {
		SendEmail(to []string, subject, body string) error
	}
)
