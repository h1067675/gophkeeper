// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран регистрации пользователя
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// registerModel хранит модель экрана регистрации пользователя
type registerModel struct {
	keys KeyMap
	help help.Model

	login textinput.Model
	pass1 textinput.Model
	pass2 textinput.Model
	email textinput.Model
	err   string
	focus int
	done  bool
	next  int
}

// newRegisterModel инициализирует модель экрана регистрации пользователя
func newRegisterModel() registerModel {
	login := textinput.New()
	login.Placeholder = "Логин"
	login.Focus()
	login.CharLimit = 32
	login.Width = 30

	email := textinput.New()
	email.Placeholder = "Email"
	email.CharLimit = 64
	email.Width = 30

	pass1 := textinput.New()
	pass1.Placeholder = "Пароль"
	pass1.CharLimit = 32
	pass1.Width = 30

	pass2 := textinput.New()
	pass2.Placeholder = "Подтверждение пароля"
	pass2.CharLimit = 32
	pass2.Width = 30

	return registerModel{
		keys:  DefaultKeyMap(),
		login: login,
		pass1: pass1,
		pass2: pass2,
		email: email,
		help:  help.New(),
	}
}

// Update обновляет данные полей и реализует основную логику экрана
func (m registerModel) Update(msg tea.Msg) (registerModel, tea.Cmd, bool) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focus++
			if m.focus > 3 {
				m.focus = 0
			}
			m.login.Blur()
			m.email.Blur()
			m.pass1.Blur()
			m.pass2.Blur()
			switch m.focus {
			case 0:
				m.login.Focus()
			case 1:
				m.email.Focus()
			case 2:
				m.pass1.Focus()
			case 3:
				m.pass2.Focus()
			}
		case "up":
			m.focus--
			if m.focus < 0 {
				m.focus = 3
			}
			m.login.Blur()
			m.email.Blur()
			m.pass1.Blur()
			m.pass2.Blur()
			switch m.focus {
			case 0:
				m.login.Focus()
			case 1:
				m.email.Focus()
			case 2:
				m.pass1.Focus()
			case 3:
				m.pass2.Focus()
			}
		case "enter":
			if copyPassword(m.pass1.Value(), m.pass2.Value()) {
				m.done = true
				return m, nil, true
			}
			m.err = "Ошибка! Пароли не соответствуют доуг другу"
		case "esc":
			m.next = -1
			return m, nil, true
		case "ctrl+c":
			return m, tea.Quit, false
		}
	}

	m.login, _ = m.login.Update(msg)
	m.email, _ = m.email.Update(msg)
	m.pass1, _ = m.pass1.Update(msg)
	m.pass2, _ = m.pass2.Update(msg)
	return m, cmd, false
}

// copyPassword проверяет совпадают ли пароли
func copyPassword(pass1, pass2 string) bool {
	return pass1 == pass2
}

// View перерисовывает экран после обновления
func (m registerModel) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keys.Tab,
		m.keys.Enter,
		m.keys.Back,
		m.keys.Quit,
	})

	title := titleStyle.Render("Введите ваш логин, электронную почту, пароль и подтверждение пароля")
	screen := borderStyle.Render(centerWidth80Style.Render(fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n\n\n\n%s\n\n",
		title,
		m.login.View(),
		m.email.View(),
		m.pass1.View(),
		m.pass2.View(),
		m.err,
	)))

	return lipgloss.JoinHorizontal(lipgloss.Top, screen) + "\n\n" + help
}
