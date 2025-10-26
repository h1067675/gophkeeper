// Package errors подменяет стандартный пакет errors для раширения отслеживания ошибок в проекте
package errors

import (
	"errors"
	"fmt"
)

// собственные ошибки пакета
var (
	// SQL
	ErrSQLNoRows = errors.New("sql: no rows in result set")
	// HTTPS Codes
	ErrStatusForbidden = errors.New("access denied")
	ErrStatusConflict  = errors.New("user with this login or email already exists")
	// repository errors
	ErrRepositoryNotInit = errors.New("repository is not initialized")
	// client errors
	ErrInvalideTransportPort = errors.New("transport port not set")
	ErrInvalideNetAddress    = errors.New("transport host or port not set")
	// server errors
	ErrInvalidContentType = errors.New("unsupported content type")
	// system errors
	ErrSystemError       = errors.New("system error")
	ErrInvalidHashFormat = errors.New("invalid hash format")
	// session errors
	ErrTOTPSessionExpired = errors.New("the one-time password has expired ")
	ErrTOTPUpdate         = errors.New("TOTP update error")
	ErrTOTPExpired        = errors.New("TOTP authorization has expired")
	ErrTOTPCodeCheck      = errors.New("TOTP code verification error")
	ErrInvalidSession     = errors.New("session is empty")
	// user errors
	ErrUserSecretPassword             = errors.New("user crypto password is not set")
	ErrUserAuthorization              = errors.New("invalide login or password")
	ErrUserAlreadyExist               = errors.New("user already exist")
	ErrUserEmailAlreadyExist          = errors.New("user email already exist")
	ErrUserNotFound                   = errors.New("user not found")
	ErrUserInvalidLogin               = errors.New("invalid login: must start with a letter, contain only Latin letters, digits, dot, dash or underscore, and be >= 6 and <= 20 chars")
	ErrUserInvalidPasswordTooShort    = errors.New("invalid password: must be at least 8 characters long")
	ErrUserInvalidPasswordNoUppercase = errors.New("invalid password: password must contain at least one uppercase letter")
	ErrUserInvalidPasswordNoLowercase = errors.New("invalid password: password must contain at least one lowercase letter")
	ErrUserInvalidPasswordNoDigits    = errors.New("invalid password: password must contain at least one digit")
	ErrUserInvalidPasswordNoSpecial   = errors.New("invalid password: password must contain at least one special character")
	ErrUserInvalidEmail               = errors.New("invalid email format")
	ErrUserInvalidPhone               = errors.New("invalid phone number format")
)

// Wrapf добавляет форматированную ошибку к ошибке
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// Is алиас для errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As алиас для errors.As
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Join алиас для errors.Join
func Join(err ...error) error {
	return errors.Join(err...)
}

// New алиас для errors.New
func New(str string) error {
	return errors.New(str)
}
