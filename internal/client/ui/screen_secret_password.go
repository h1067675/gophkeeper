// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран установки пароля шифрования
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// secretModel хранит модель экрана установки пароля шифрования
type secretModel struct {
	keys KeyMap
	help help.Model

	password textinput.Model
	repeat   textinput.Model
	create   bool
	focus    int
	err      string
}

// newsecretModel инициализирует модель экрана установки пароля шифрования
func newsecretModel() *secretModel {
	pass1 := textinput.New()
	pass1.Placeholder = "Пароль"
	pass1.Focus()
	pass1.CharLimit = 32
	pass1.Width = 30

	pass2 := textinput.New()
	pass2.Placeholder = "Повтор пароля"
	pass2.CharLimit = 32
	pass2.Width = 30

	return &secretModel{
		password: pass1,
		repeat:   pass2,
		keys:     DefaultKeyMap(),
		help:     help.New(),
	}
}

// Update обновляет данные полей и реализует основную логику экрана
func (m secretModel) Update(msg tea.Msg) (secretModel, tea.Cmd, bool) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Tab):
			m.focus++
			if m.focus > 1 {
				m.focus = 0
			}
			if m.focus == 0 {
				m.password.Focus()
				m.repeat.Blur()
			} else {
				m.repeat.Focus()
				m.password.Blur()
			}
		case key.Matches(msg, m.keys.Enter):
			if m.password.Value() == "" {
				m.err = "Ошибка! Пароль не может быть пустым"
				return m, nil, false
			}
			if m.password.Value() != m.repeat.Value() && m.create == true {
				m.err = "Ошибка! Пароли не совпадают"
				return m, nil, false
			}
			return m, nil, true
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit, false
		}
	}

	m.password, _ = m.password.Update(msg)
	m.repeat, _ = m.repeat.Update(msg)
	return m, cmd, false
}

// View перерисовывает экран после обновления
func (m secretModel) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keys.Tab,
		m.keys.Enter,
		m.keys.Back,
		m.keys.Quit,
	})

	title := titleStyle.Render("Введите Ваш пароль шифрования")
	title1, repeat := "", ""
	if m.create {
		title1 = textStyle.Render(fmt.Sprint("Внимание! Вы создаете пароль шифрования.\n",
			"Данный пароль не хранится ни на Вашем устройстве ни на сервере.\n",
			"Обязательно запишите его или надежно запомните, в случае его \nутери восстановить данные будет невозможно"))
		repeat = m.repeat.View()
	}
	screen := borderStyle.Render(centerWidth80Style.Render(fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n\n\n\n\n%s\n\n",
		title,
		title1,
		m.password.View(),
		repeat,
		m.err,
	)))

	return lipgloss.JoinHorizontal(lipgloss.Top, screen) + "\n\n" + help
}
