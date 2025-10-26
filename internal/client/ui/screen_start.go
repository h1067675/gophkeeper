// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает стартовый экран
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// startModel хранит модель стартового экрана
type startModel struct {
	keys KeyMap
	help help.Model

	items        []item
	selected     int
	addingServer bool
	hasConfig    bool
	sessionOk    bool
	err          string
	make         int
	done         bool

	buildVersion string
	buildDate    string
	buildCommit  string
}

// newStartModel инициализирует модель стартового экрана
func newStartModel() startModel {

	return startModel{
		keys: DefaultKeyMap(),
		help: help.New(),
	}
}

// Init устанавливает поля меню в зависимости от входящих данных
func (m *startModel) Init() tea.Cmd {
	if m.sessionOk && m.hasConfig {
		m.items = append(m.items, item{title: "Подключиться к серверу с существующей сессией", make: 1})
		m.items = append(m.items, item{title: "Сбросить сессию и войти как новый пользователь", make: 2})
	} else if m.hasConfig {
		m.items = append(m.items, item{title: "Подключиться к серверу с сохраненными ранее настройками", make: 3})
	} else {
		m.items = append(m.items, item{title: "Подключиться к тестовому серверу localhost", make: 0})
	}
	m.items = append(m.items, item{title: "Изменить настройки сервера", make: 4})
	m.items = append(m.items, item{title: "Выйти", make: -1})

	return nil
}

// Update обновляет данные полей и реализует основную логику экрана
func (m startModel) Update(msg tea.Msg) (startModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down) || key.Matches(msg, m.keys.Tab):
			m.selected++
			if m.selected >= len(m.items) {
				m.selected = 0
			}
			return m, nil, false
		case key.Matches(msg, m.keys.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = len(m.items) - 1
			}
			return m, nil, false
		case key.Matches(msg, m.keys.Enter):
			m.make = m.items[m.selected].make
			return m, nil, true
		case key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.Back):
			m.make = -1
			return m, tea.Quit, false
		}
	}
	return m, nil, false
}

// View перерисовывает экран после обновления
func (m startModel) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keys.Up,
		m.keys.Down,
		m.keys.Enter,
		m.keys.Quit,
	})

	if m.make == -1 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("\nGoodbye 👋\n")
	}
	title := titleStyle.Render("Добро пожаловать в GophKeeper")
	subtitle := textStyle.Render("Это защищенное хранилище для ваших секретов 🔐")
	items := ""
	for i := 0; i < 4; i++ {
		if i == m.selected {
			items += selectedItemStyle.Render(m.items[i].title)
		} else if i < len(m.items) {
			items += itemStyle.Render(m.items[i].title)
		}
		items += "\n"
	}
	options := fmt.Sprintf(
		"\n\n%s\n%s\n%s",
		items,
		errorStyle.Render(m.err),
		versionsStyle.Render("Version:", m.buildVersion, "Date:", m.buildDate, "Commit:", m.buildCommit),
	)

	screen := borderStyle.Render(centerWidth80Style.Render(fmt.Sprintf("%s\n\n%s\n%s",
		title,
		subtitle,
		options,
	)))
	return lipgloss.JoinHorizontal(lipgloss.Top, screen) + "\n\n" + help
}
