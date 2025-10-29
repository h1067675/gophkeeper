package ui

import (
	"gophkeeper/pkg/domain"

	tea "github.com/charmbracelet/bubbletea"
)

// StartScreenDone реализует логику стартового экрана TUI
func (m model) StartScreenDone(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch m.startScreen.make {
	case -1:
		return m, tea.Quit
	case 0:
		err := m.service.SaveServerConfiguration(domain.LocalServerAddress, domain.LocalPortGRPC, domain.LocalPortHTTPS)
		if err != nil {
			m.startScreen.err = "Ошибка! Настройки не сохранены, проверьте их правильность."
			break
		}
		m.currentScreen = domain.ScreenLogin
	case 1:
		m.currentScreen = domain.ScreenTOTP
	case 2:
		if err := m.service.SessionExit(); err != "" {
			m.startScreen.err = err
		}
		fallthrough
	case 3:
		m.currentScreen = domain.ScreenLogin
	case 4:
		m.currentScreen = domain.ScreenConfig
	}
	m.loadNextScreen(m.currentScreen)
	return m, cmd
}

// ConfifScreenDone реализует логику экрана конфигурации TUI
func (m model) ConfifScreenDone(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	// если нажали назад уходим на стартовый экран
	if m.configScreen.next > -1 {
		m.loadNextScreen(m.configScreen.next)
		return m, cmd
	}
	if err := m.service.SaveServerConfiguration(m.configScreen.server.Value(), m.configScreen.GRPCport.Value(), m.configScreen.HTTPSport.Value()); err != nil {
		m.configScreen.err = "Ошибка! Настройки не сохранены, проверьте правильность их указания."
		return m, cmd
	}
	m.loadNextScreen(m.configScreen.next)
	return m, cmd
}

// LoginScreenDone реализует логику экрана авторизации TUI
func (m model) LoginScreenDone(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if m.loginScreen.next > -1 {
		m.loadNextScreen(m.loginScreen.next)
		return m, cmd
	}
	m.service.User.Login, m.service.User.Password = m.loginScreen.login.Value(), m.loginScreen.pass.Value()
	nextScreen, err := m.service.Authorization()
	if err != "" {
		m.loginScreen.err = err
	}
	m.loadNextScreen(nextScreen)
	return m, cmd
}

// RegisterScreenDone реализует логику экрана авторизации TUI
func (m model) RegisterScreenDone(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if m.registerScreen.next > -1 {
		m.loadNextScreen(m.registerScreen.next)
		return m, cmd
	}
	m.service.User.Login, m.service.User.Password, m.service.User.Email = m.registerScreen.login.Value(), m.registerScreen.pass1.Value(), m.registerScreen.email.Value()
	nextScreen, err := m.service.Registration()
	if err != "" {
		m.registerScreen.err = err
		return m, cmd
	}
	m.loadNextScreen(nextScreen)
	if m.service.User.TOTP != "" {
		m.TOTPScreen.qr = m.service.QRMediumToString(m.service.User.TOTP)
	}
	return m, cmd
}

// TOTPScreenDone реализует логику экрана двухфакторной авторизации TUI
func (m model) TOTPScreenDone(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	nextScreen := domain.ScreenTOTP
	switch m.TOTPScreen.make {
	case 0:
		m.TOTPScreen.showQR = false
	case 1:
		nextScreen, m.TOTPScreen.err = m.service.TOTPConfirmation(m.TOTPScreen.code.Value())
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
	if nextScreen != domain.ScreenTOTP {
		m.loadNextScreen(nextScreen)
		if nextScreen == domain.ScreenSecretPassword {
			if m.service.User.SPassHash == "" {
				m.secretScreen.create = true
			}
		}
	}
	return m, cmd
}

// SecretScreenDone реализует логику экрана ввода пароля шифрования TUI
func (m model) SecretScreenDone(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	next, err := domain.ScreenSecretPassword, ""
	m.service.User.ClientPass = m.secretScreen.password.Value()
	if m.secretScreen.create {
		next, err = m.service.SaveHashSecretPassword()
	} else {
		next, err = m.service.CheckSecretPassword()
	}
	m.secretScreen.err = err
	if next == domain.ScreenHome {
		m.loadNextScreen(next)
		terr, _ := m.service.SyncData()
		m.homeScreen.err += terr
		return m, cmd
	}
	return m, cmd
}

// HomeScreenDone реализует логику экрана выбора редактора TUI
func (m model) HomeScreenDone(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch m.homeScreen.make {
	case -1:
		return m, tea.Quit
	case 1:
		data, err := m.service.GetPasswordsList()
		m.currentScreen = domain.ScreenPasswords
		m.loadNextScreen(m.currentScreen)
		m.passwordsScreen.err = err
		m.passwordsScreen.data = data
		m.passwordsScreen.Init()
	case 2:
		cards, err := m.service.GetCardsList()
		m.currentScreen = domain.ScreenCards
		m.loadNextScreen(m.currentScreen)
		m.cardsScreen.err = err
		m.cardsScreen.cards = cards
		m.cardsScreen.Init()
	case 3:
		data, err := m.service.GetTextDataList()
		m.currentScreen = domain.ScreenTexts
		m.loadNextScreen(m.currentScreen)
		m.textScreen.err = err
		m.textScreen.data = data
		m.textScreen.Init()
	case 4:
		data, err := m.service.GetBinaryDataList()
		m.currentScreen = domain.ScreenBinaries
		m.loadNextScreen(m.currentScreen)
		m.binaryScreen.err = err
		m.binaryScreen.data = data
		m.binaryScreen.Init()
	case 5:
		err, _ := m.service.SyncData()
		m.homeScreen.err = err
	case 8:
		m.currentScreen = domain.ScreenStart
		m.loadNextScreen(m.currentScreen)
	}
	return m, cmd
}

// PasswordScreenAction реализует логику действий редактора паролей
func (m model) PasswordScreenAction(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch m.passwordsScreen.action {
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
}

// CardsScreenAction реализует логику действий редактора банковских карт
func (m model) CardsScreenAction(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch m.cardsScreen.action {
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
}

// TextScreenAction реализует логику действий редактора текстов
func (m model) TextScreenAction(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch m.textScreen.action {
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
}

// BinaryScreenAction реализует логику действий редактора двоичных данных
func (m model) BinaryScreenAction(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch m.binaryScreen.action {
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
