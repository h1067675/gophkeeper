// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран конфигурации сервера приложения
package ui

import (
	"fmt"
	"gophkeeper/pkg/domain"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// cardsModel хранит модель экран работы с данными банковских карт пользователя
type configModel struct {
	keys KeyMap
	help help.Model

	server    textinput.Model
	GRPCport  textinput.Model
	HTTPSport textinput.Model
	next      domain.Screen
	focus     int
	done      bool
	err       string
	selected  string
}

// newConfigModel инициализирует модель экран работы с данными банковских карт пользователя
func newConfigModel() configModel {
	server := textinput.New()
	server.Placeholder = "Сервер"
	server.Focus()
	server.CharLimit = 32
	server.Width = 30

	g := textinput.New()
	g.Placeholder = "gRPC порт"
	g.CharLimit = 32
	g.Width = 30

	h := textinput.New()
	h.Placeholder = "HTTPS порт"
	h.CharLimit = 32
	h.Width = 30

	return configModel{
		keys:      DefaultKeyMap(),
		server:    server,
		GRPCport:  g,
		HTTPSport: h,
		help:      help.New(),
	}
}

// Update обновляет данные полей и реализует основную логику экрана
func (m configModel) Update(msg tea.Msg) (configModel, tea.Cmd, bool) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Tab):
			m.focus++
			if m.focus > 2 {
				m.focus = 0
			}
			m.server.Blur()
			m.GRPCport.Blur()
			m.HTTPSport.Blur()
			switch m.focus {
			case 0:
				m.server.Focus()
			case 1:
				m.GRPCport.Focus()
			case 2:
				m.HTTPSport.Focus()
			}
		case key.Matches(msg, m.keys.Back):
			m.next = domain.ScreenStart
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit, false
		case key.Matches(msg, m.keys.Enter):
			m.next = domain.ScreenLogin
			if m.server.Value() == "" {
				m.err = "Ошибка! Адрес сервера не может быть пустым"
				return m, nil, false
			}
			if m.GRPCport.Value() == "" && m.HTTPSport.Value() == "" {
				m.err = "Ошибка! Один из портов обязательно должен быть указан"
				return m, nil, false
			}
			return m, nil, true
		}
	}
	m.server, _ = m.server.Update(msg)
	m.HTTPSport, _ = m.HTTPSport.Update(msg)
	m.GRPCport, _ = m.GRPCport.Update(msg)
	return m, cmd, false
}

// View перерисовывает экран после обновления
func (m configModel) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keys.Tab,
		m.keys.Enter,
		m.keys.Back,
		m.keys.Quit,
	})

	title := titleStyle.Render("Введите адрес и порт сервера GophKeeper")
	screen := borderStyle.Render(centerWidth80Style.Render(fmt.Sprintf("%s\n\n\n%s\n%s\n%s\n\n\n\n%s\n\n",
		title,
		m.server.View(),
		m.GRPCport.View(),
		m.HTTPSport.View(),
		m.err,
	)))

	return lipgloss.JoinHorizontal(lipgloss.Top, screen) + "\n\n" + help
}
