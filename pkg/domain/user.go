// Package domain описывает форматы сущностей всего проекта
package domain

import (
	"gophkeeper/pkg/errors"
	"regexp"
)

// User сущность пользователя
type User struct {
	Login      string
	Password   string
	ClientPass string
	SPassHash  string
	Email      string
	Session    string
	Token      string
	ID         int
	TOTP       string
	SaltB64    string
}

// Validate проверяет верность заполнения данных
func (u *User) Validate() error {
	return errors.Join(
		ValidateLogin(u.Login),
		ValidatePassword(u.Password),
		ValidateEmail(u.Email),
	)
}

// ValidateLogin проверяет верность указания логина
func ValidateLogin(login string) error {
	// Логин должен:
	// - начинается с буквы
	// - быть от 6 до 20 символов
	// и может содержать:
	// - латинские буквы
	// - цифры
	// - символы: ., -, _
	var loginRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]{4,19}$`)
	if !loginRegexp.MatchString(login) {
		return errors.ErrUserInvalidLogin
	}
	return nil
}

// ValidatePassword проверяет верность указания пароля
func ValidatePassword(password string) error {
	// Пароль должен:
	// - быть минимум 8 символов
	// - иметь хотя бы одну заглавную букву из латинского алфивита, одну цифру и один спецсимвол
	if len(password) < 8 {
		return errors.ErrUserInvalidPasswordTooShort
	}
	upper := regexp.MustCompile(`[A-Z]`)
	if !upper.MatchString(password) {
		return errors.ErrUserInvalidPasswordNoUppercase
	}
	lower := regexp.MustCompile(`[a-z]`)
	if !lower.MatchString(password) {
		return errors.ErrUserInvalidPasswordNoLowercase
	}
	digit := regexp.MustCompile(`\d`)
	if !digit.MatchString(password) {
		return errors.ErrUserInvalidPasswordNoDigits
	}
	special := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`)
	if !special.MatchString(password) {
		return errors.ErrUserInvalidPasswordNoSpecial
	}
	return nil
}

// ValidateEmail проверяет верность указания 'электронной почты'
func ValidateEmail(email string) error {
	var emailRegexp = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	if !emailRegexp.MatchString(email) {
		return errors.ErrUserInvalidEmail
	}
	return nil
}
