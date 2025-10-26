// Package authorization реализует функции авторизации пользователя по сессиям и токенам, а также функцию двухфакторной авторизации
package authorization

import (
	"gophkeeper/pkg/errors"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// GenerateTOTPSecret генерирует данные доступа в Google Auth по аредоставленным данным пользователя
func (a *Autorization) GenerateTOTPSecret(issuer string, email string) (key *otp.Key, err error) {
	key, err = totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ValidateTOTPCode сверяет полученный код отпользователя с кодом предоставленным Goggle Auth
func (a *Autorization) ValidateTOTPCode(code string, secretKey string) error {
	if !totp.Validate(code, secretKey) {
		return errors.ErrTOTPCodeCheck
	}
	return nil
}
