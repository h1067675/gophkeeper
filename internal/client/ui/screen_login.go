// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран авторизации пользователя
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// loginModel хранит модель экрана авторизации пользователя
type loginModel struct {
	keys KeyMap
	help help.Model

	login      textinput.Model
	pass       textinput.Model
	err        string
	focus      int
	done       bool
	toRegister bool
}

// newLoginModel инициализирует модель экрана авторизации пользователя
func newLoginModel() loginModel {
	login := textinput.New()
	login.Placeholder = "Имя пользователя"
	login.Focus()
	login.CharLimit = 32
	login.Width = 30

	pass := textinput.New()
	pass.Placeholder = "Пароль"
	pass.CharLimit = 32
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '•'
	pass.Width = 30

	return loginModel{
		login: login,
		pass:  pass,
		keys:  DefaultKeyMap(),
		help:  help.New(),
	}
}

// Update обновляет данные полей и реализует основную логику экрана
func (m loginModel) Update(msg tea.Msg) (loginModel, tea.Cmd, bool, bool) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focus++
			if m.focus > 1 {
				m.focus = 0
			}
			if m.focus == 0 {
				m.login.Focus()
				m.pass.Blur()
			} else {
				m.pass.Focus()
				m.login.Blur()
			}
		case "ctrl+r":
			m.toRegister = true
			return m, nil, false, true
		case "enter":
			if m.login.Value() == "" {
				m.err = "Ошибка! Логин не может быть пустым"
				return m, nil, false, false
			}
			if m.pass.Value() == "" {
				m.err = "Ошибка! Пароль не может быть пустым"
				return m, nil, false, false
			}
			m.done = true
			return m, nil, true, false
		case "esc", "ctrl+c":
			return m, tea.Quit, false, false
		}
	}

	m.login, _ = m.login.Update(msg)
	m.pass, _ = m.pass.Update(msg)
	return m, cmd, false, false
}

// View перерисовывает экран после обновления
func (m loginModel) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keys.Tab,
		m.keys.Register,
		m.keys.Enter,
		m.keys.Back,
		m.keys.Quit,
	})

	title := titleStyle.Render("Введите ваши логин и пароль или перейдите к регистрации")
	screen := borderStyle.Render(centerWidth80Style.Render(fmt.Sprintf("%s\n\n\n%s\n%s\n\n\n\n\n%s\n\n",
		title,
		m.login.View(),
		m.pass.View(),
		m.err,
	)))

	return lipgloss.JoinHorizontal(lipgloss.Top, screen) + "\n\n" + help
}
