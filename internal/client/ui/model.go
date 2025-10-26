// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// основная бизнес логика графического интерфейса
package ui

import (
	"gophkeeper/internal/client/gophkeeper"
	"gophkeeper/pkg/domain"

	tea "github.com/charmbracelet/bubbletea"
)

// model описывает основную модель графического интерфейса
type model struct {
	service *gophkeeper.Service

	currentScreen   domain.Screen
	startScreen     startModel
	configScreen    configModel
	loginScreen     loginModel
	registerScreen  registerModel
	TOTPScreen      totpModel
	secretScreen    secretModel
	homeScreen      homeModel
	passwordsScreen passwordsModel
	cardsScreen     cardsModel
	textScreen      textModel
	binaryScreen    binaryModel
}

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
		m.startScreen = screen
		if done {
			switch screen.make {
			case -1:
				return m, tea.Quit
			case 0:
				err := m.service.SaveServerConficuration("https://localhost", "", "8853")
				if err != nil {
					m.startScreen.err = "Ошибка! Настройки не сохранены, проверьте их правильность."
					return m, cmd
				}
				m.loginScreen = newLoginModel()
				m.currentScreen = domain.ScreenLogin
			case 1:
				m.TOTPScreen = newTOTPModel("")
				m.currentScreen = domain.ScreenTOTP
			case 2:
				if err := m.service.SessionExit(); err != "" {
					m.startScreen.err = err
				}
				fallthrough
			case 3:
				m.loginScreen = newLoginModel()
				m.currentScreen = domain.ScreenLogin
			case 4:
				m.configScreen = newConfigModel()
				m.currentScreen = domain.ScreenConfig
			}
		}
		return m, cmd

	case domain.ScreenConfig:
		screen, cmd, done := m.configScreen.Update(msg)
		m.configScreen = screen
		if done {
			if m.service.Configurer.CheckNetAddress(m.configScreen.server.Value()+":"+m.configScreen.HTTPSport.Value()) != nil && m.service.Configurer.CheckNetAddress(m.configScreen.server.Value()+":"+m.configScreen.GRPCport.Value()) != nil {
				m.configScreen.err = "Ошибка! Некорректно указаны адреса сервера"
				return m, cmd
			}
			err := m.service.SaveServerConficuration(m.configScreen.server.Value(), m.configScreen.GRPCport.Value(), m.configScreen.HTTPSport.Value())
			if err != nil {
				m.configScreen.err = "Ошибка! Настройки не сохранены, проверьте правильность настройки."
				return m, cmd
			}
		}
		switch screen.next {
		case domain.ScreenLogin:
			m.loginScreen = newLoginModel()
			m.currentScreen = domain.ScreenLogin
		case domain.ScreenStart:
			m.startScreen = newStartModel()
			m.currentScreen = domain.ScreenStart
		}
		return m, cmd

	case domain.ScreenLogin:
		screen, cmd, done, toRegister := m.loginScreen.Update(msg)
		m.loginScreen = screen
		if toRegister {
			m.registerScreen = newRegisterModel()
			m.currentScreen = domain.ScreenRegister
		}
		if done {
			if m.loginScreen.login.Value() == "" || m.loginScreen.pass.Value() == "" {
				m.loginScreen.err = "Ошибка! Логин и пароль не могут быть пустыми."
				return m, cmd
			}
			m.service.User.Login, m.service.User.Password = m.loginScreen.login.Value(), m.loginScreen.pass.Value()
			nextScreen, err := m.service.Authorization()
			if err != "" {
				m.loginScreen.err = err
			}
			m.TOTPScreen = newTOTPModel("")
			m.currentScreen = nextScreen
		}
		return m, cmd

	case domain.ScreenRegister:
		screen, cmd, done := m.registerScreen.Update(msg)
		m.registerScreen = screen
		if done {
			if m.registerScreen.next == -1 {
				m.currentScreen = domain.ScreenLogin
				return m, cmd
			}
			m.service.User.Login, m.service.User.Password, m.service.User.Email = m.registerScreen.login.Value(), m.registerScreen.pass1.Value(), m.registerScreen.email.Value()
			nextScreen, err := m.service.Registration()
			if m.service.User.TOTP != "" {
				m.TOTPScreen = newTOTPModel(m.service.QRMediumToString(m.service.User.TOTP))
			} else {
				m.TOTPScreen = newTOTPModel("")
			}
			if err != "" {
				m.registerScreen.err = err
			}
			m.currentScreen = nextScreen
		}
		return m, cmd

	case domain.ScreenTOTP:
		screen, cmd, done := m.TOTPScreen.Update(msg)
		m.TOTPScreen = screen
		nextScreen := domain.ScreenTOTP
		if done {
			switch m.TOTPScreen.make {
			case 0:
				m.TOTPScreen.showQR = false
			case 1:
				nextScreen, m.TOTPScreen.err = m.service.TOTPConfirmation(m.TOTPScreen.code.Value())
				m.secretScreen = newsecretModel()
				if m.service.User.SPassHash == "" {
					m.secretScreen.create = true
				}
			case 3:
				m.service.User.Email = m.TOTPScreen.email.Value()
				nextScreen, m.TOTPScreen.err = m.service.TOTPUpdate(m.TOTPScreen.code.Value())
			case 4:
				nextScreen, m.TOTPScreen.err = m.service.TOTPUpdate(m.TOTPScreen.code.Value())
				m.TOTPScreen.code.SetValue("")
				if m.service.User.TOTP != "" {
					m.TOTPScreen.qr = m.service.QRMediumToString(m.service.User.TOTP)
				}
			case 5:
				m.TOTPScreen.showQR = true
				m.TOTPScreen.make = 0
			case 6:
				nextScreen = domain.ScreenStart
			}
		}
		m.currentScreen = nextScreen
		return m, cmd

	case domain.ScreenSecretPassword:
		screen, cmd, done := m.secretScreen.Update(msg)
		m.secretScreen = screen
		next, err := domain.ScreenSecretPassword, ""
		if done {
			m.service.User.ClientPass = screen.password.Value()
			if screen.create {
				next, err = m.service.SaveHashSecretPassword()
			} else {
				next, err = m.service.CheckSecretPassword()
			}
			m.homeScreen = newHomeModel()
			m.secretScreen.err = err
			if next == domain.ScreenHome {
				terr, _ := m.service.SyncData()
				m.homeScreen.err += terr
			}
			m.currentScreen = next
		}
		return m, cmd

	case domain.ScreenHome:
		m.configScreen = configModel{}
		m.registerScreen = registerModel{}
		m.loginScreen = loginModel{}
		m.TOTPScreen = totpModel{}
		m.secretScreen = secretModel{}

		screen, cmd, done := m.homeScreen.Update(msg)
		m.homeScreen = screen
		if done {
			switch screen.make {
			case -1:
				return m, tea.Quit
			case 1:
				m.passwordsScreen = newPasswordsModel()
				data, err := m.service.GetPasswordsList()
				m.passwordsScreen.err = err
				m.passwordsScreen.data = data
				m.passwordsScreen.Init()
				m.currentScreen = domain.ScreenPasswords
			case 2:
				m.cardsScreen = newCardsModel()
				cards, err := m.service.GetCardsList()
				m.cardsScreen.err = err
				m.cardsScreen.cards = cards
				m.cardsScreen.Init()
				m.currentScreen = domain.ScreenCards
			case 3:
				m.textScreen = newTextModel()
				data, err := m.service.GetTextDataList()
				m.textScreen.err = err
				m.textScreen.data = data
				m.textScreen.Init()
				m.currentScreen = domain.ScreenTexts
			case 4:
				m.binaryScreen = newBinaryModel()
				data, err := m.service.GetBinaryDataList()
				m.binaryScreen.err = err
				m.binaryScreen.width = 6
				m.binaryScreen.data = data
				m.binaryScreen.Init()
				m.currentScreen = domain.ScreenBinaries
			case 5:
				terr, _ := m.service.SyncData()
				m.homeScreen.err = terr
			case 8:
				m.currentScreen = domain.ScreenStart
			}
		}
		return m, cmd

	case domain.ScreenPasswords:
		screen, cmd := m.passwordsScreen.Update(msg)
		m.passwordsScreen = screen
		switch screen.action {
		case save:
			m.passwordsScreen.err = m.service.SavePassword(&m.passwordsScreen.actionData)
		case delete:
			m.passwordsScreen.err = m.service.DeletePassword(m.passwordsScreen.actionData)
		case load:
			m.passwordsScreen.actionData, m.passwordsScreen.err = m.service.GetPassword(m.passwordsScreen.actionData.LocalID)
			m.passwordsScreen.setDataToRightPanel()
		case back:
			m.currentScreen = domain.ScreenHome
		}
		m.passwordsScreen.action = none
		return m, cmd

	case domain.ScreenCards:
		screen, cmd := m.cardsScreen.Update(msg)
		m.cardsScreen = screen
		switch screen.action {
		case save:
			m.cardsScreen.err = m.service.SaveCard(&m.cardsScreen.actionCard)
			m.cardsScreen.setDataToCard()
		case delete:
			m.cardsScreen.err = m.service.DeleteCard(m.cardsScreen.actionCard)
		case load:
			m.cardsScreen.actionCard, m.cardsScreen.err = m.service.GetCard(m.cardsScreen.actionCard.LocalID)
			m.cardsScreen.setDataToCard()
		case back:
			m.currentScreen = domain.ScreenHome
		}
		m.cardsScreen.action = none
		return m, cmd

	case domain.ScreenTexts:
		screen, cmd := m.textScreen.Update(msg)
		m.textScreen = screen
		switch screen.action {
		case save:
			m.textScreen.err = m.service.SaveTextData(&m.textScreen.actionData)
			m.textScreen.setDataToRightPanel()
		case delete:
			m.textScreen.err = m.service.DeleteTextData(m.textScreen.actionData)
		case load:
			m.textScreen.actionData, m.textScreen.err = m.service.GetTextData(m.textScreen.actionData.LocalID)
			m.textScreen.setDataToRightPanel()
		case back:
			m.currentScreen = domain.ScreenHome
		}
		m.textScreen.action = none
		return m, cmd

	case domain.ScreenBinaries:
		screen, cmd := m.binaryScreen.Update(msg)
		m.binaryScreen = screen
		switch screen.action {
		case save:
			m.binaryScreen.err = m.service.SaveBinaryData(&m.binaryScreen.actionData)
			m.binaryScreen.setDataToRightPanel()
		case delete:
			m.binaryScreen.err = m.service.DeleteBinaryData(m.binaryScreen.actionData)
		case load:
			m.binaryScreen.actionData, m.binaryScreen.err = m.service.GetBinaryData(m.binaryScreen.actionData.LocalID)
			m.binaryScreen.setDataToRightPanel()
		case back:
			m.currentScreen = domain.ScreenHome
		}
		m.binaryScreen.action = none
		return m, cmd
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
