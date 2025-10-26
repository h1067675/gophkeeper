// Package domain описывает форматы сущностей всего проекта
package domain

// Описывает типы экранов графического интерфейса
const (
	ScreenStart Screen = iota
	ScreenConfig
	ScreenLogin
	ScreenRegister
	ScreenTOTP
	ScreenHome
	ScreenSecretPassword
	ScreenPasswords
	ScreenCards
	ScreenTexts
	ScreenBinaries
)

// Screen необходима для описания типов экранов графического интерфейса
type Screen int
