// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// основная бизнес логика графического интерфейса
package ui

import (
	"gophkeeper/internal/client/gophkeeper"
	"gophkeeper/pkg/domain"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	// model описывает основную модель графического интерфейса
	model struct {
		service *gophkeeper.Service

		currentScreen   domain.Screen
		startScreen     *startModel
		configScreen    *configModel
		loginScreen     *loginModel
		registerScreen  *registerModel
		TOTPScreen      *totpModel
		secretScreen    *secretModel
		homeScreen      *homeModel
		passwordsScreen *passwordsModel
		cardsScreen     *cardsModel
		textScreen      *textModel
		binaryScreen    *binaryModel
	}
)

// NewModel инициализирует основную модель гоафического интерфейса
func NewModel(service *gophkeeper.Service) model {
	screen := newStartModel()
	screen.buildVersion = service.Versions.BuildVersion
	screen.buildDate = service.Versions.BuildDate
	screen.buildCommit = service.Versions.BuildCommit

	if service.User.Session != "" {
		screen.sessionOk = true
	}
	if err := service.Configurer.CheckConfig(); err == nil {
		screen.hasConfig = true
	}
	screen.Init()
	return model{
		currentScreen: domain.ScreenStart,
		startScreen:   screen,
		service:       service,
	}
}

// Init функция
func (m model) Init() tea.Cmd { return nil }

// Update обновляет данные экранов и релизует логику перемещения между ними, а также осуществляет связь с основным сервисом
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentScreen {

	case domain.ScreenStart:
		screen, cmd, done := m.startScreen.Update(msg)
		m.startScreen = &screen
		if done {
			return m.StartScreenDone(cmd)
		}
		return m, cmd

	case domain.ScreenConfig:
		screen, cmd, done := m.configScreen.Update(msg)
		m.configScreen = &screen
		if done {
			return m.ConfifScreenDone(cmd)
		}
		return m, cmd

	case domain.ScreenLogin:
		screen, cmd, done := m.loginScreen.Update(msg)
		m.loginScreen = &screen
		if done {
			return m.LoginScreenDone(cmd)
		}
		return m, cmd

	case domain.ScreenRegister:
		screen, cmd, done := m.registerScreen.Update(msg)
		m.registerScreen = &screen
		if done {
			return m.RegisterScreenDone(cmd)
		}
		return m, cmd

	case domain.ScreenTOTP:
		screen, cmd, done := m.TOTPScreen.Update(msg)
		m.TOTPScreen = &screen
		if done {
			return m.TOTPScreenDone(cmd)
		}
		return m, cmd

	case domain.ScreenSecretPassword:
		screen, cmd, done := m.secretScreen.Update(msg)
		m.secretScreen = &screen
		if done {
			return m.SecretScreenDone(cmd)
		}
		return m, cmd

	case domain.ScreenHome:
		screen, cmd, done := m.homeScreen.Update(msg)
		m.homeScreen = &screen
		if done {
			return m.HomeScreenDone(cmd)
		}
		return m, cmd

	case domain.ScreenPasswords:
		screen, cmd := m.passwordsScreen.Update(msg)
		m.passwordsScreen = &screen
		return m.PasswordScreenAction(cmd)

	case domain.ScreenCards:
		screen, cmd := m.cardsScreen.Update(msg)
		m.cardsScreen = &screen
		return m.CardsScreenAction(cmd)

	case domain.ScreenTexts:
		screen, cmd := m.textScreen.Update(msg)
		m.textScreen = &screen
		return m.TextScreenAction(cmd)

	case domain.ScreenBinaries:
		screen, cmd := m.binaryScreen.Update(msg)
		m.binaryScreen = &screen
		return m.BinaryScreenAction(cmd)
	}

	return m, nil
}

// View перерисовывает экран после обновления
func (m model) View() string {
	switch m.currentScreen {
	case domain.ScreenStart:
		return m.startScreen.View()
	case domain.ScreenConfig:
		return m.configScreen.View()
	case domain.ScreenLogin:
		return m.loginScreen.View()
	case domain.ScreenRegister:
		return m.registerScreen.View()
	case domain.ScreenTOTP:
		return m.TOTPScreen.View()
	case domain.ScreenHome:
		return m.homeScreen.View()
	case domain.ScreenSecretPassword:
		return m.secretScreen.View()
	case domain.ScreenPasswords:
		return m.passwordsScreen.View()
	case domain.ScreenCards:
		return m.cardsScreen.View()
	case domain.ScreenTexts:
		return m.textScreen.View()
	case domain.ScreenBinaries:
		return m.binaryScreen.View()
	default:
		return "Unknown screen"
	}
}

// loadNextScreen загружает следующий экран TUI
func (m *model) loadNextScreen(next domain.Screen) {
	switch next {
	case domain.ScreenLogin:
		if m.loginScreen == nil {
			m.loginScreen = newLoginModel()
		}
		m.currentScreen = domain.ScreenLogin
	case domain.ScreenStart:
		if m.startScreen == nil {
			m.startScreen = newStartModel()
		}
		m.currentScreen = domain.ScreenStart
	case domain.ScreenConfig:
		if m.configScreen == nil {
			m.configScreen = newConfigModel()
		}
		m.currentScreen = domain.ScreenConfig
	case domain.ScreenTOTP:
		if m.TOTPScreen == nil {
			m.TOTPScreen = newTOTPModel()
		}
		m.currentScreen = domain.ScreenTOTP
	case domain.ScreenSecretPassword:
		if m.secretScreen == nil {
			m.secretScreen = newsecretModel()
		}
		m.currentScreen = domain.ScreenSecretPassword
	case domain.ScreenHome:
		if m.homeScreen == nil {
			m.homeScreen = newHomeModel()
		}
		m.currentScreen = domain.ScreenHome
	case domain.ScreenPasswords:
		if m.passwordsScreen == nil {
			m.passwordsScreen = newPasswordsModel()
		}
		m.currentScreen = domain.ScreenPasswords
	case domain.ScreenCards:
		if m.cardsScreen == nil {
			m.cardsScreen = newCardsModel()
		}
		m.currentScreen = domain.ScreenCards
	case domain.ScreenTexts:
		if m.textScreen == nil {
			m.textScreen = newTextModel()
		}
		m.currentScreen = domain.ScreenTexts
	case domain.ScreenBinaries:
		if m.binaryScreen == nil {
			m.binaryScreen = newBinaryModel()
		}
		m.currentScreen = domain.ScreenBinaries
	}
}
