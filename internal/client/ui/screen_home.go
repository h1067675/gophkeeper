// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран основного меню приложения на данном экране можно
// выбрать с какими данными работать и сделать синхронизацию данных
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// homeModel хранит модель экрана основного меню приложения на данном экране можно
type homeModel struct {
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
}

// newHomeModel инициализирует модель экрана основного меню приложения на данном экране можно
func newHomeModel() *homeModel {
	items := make([]item, 0)
	items = append(items, item{title: "Пароли", make: 1})
	items = append(items, item{title: "Банковские карты", make: 2})
	items = append(items, item{title: "Текстовая информация", make: 3})
	items = append(items, item{title: "Бинарные данные", make: 4})
	items = append(items, item{title: "Синхронизировать данные", make: 5})
	items = append(items, item{title: "Выйти", make: -1})
	// items = append(items, item{title: "Удалить аккаунт", make: 9})

	return &homeModel{
		keys:  DefaultKeyMap(),
		items: items,
		help:  help.New(),
	}
}

// Update обновляет данные полей и реализует основную логику экрана
func (m homeModel) Update(msg tea.Msg) (homeModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = ""
		switch {
		case key.Matches(msg, m.keys.Down) || key.Matches(msg, m.keys.Tab):
			m.selected++
			if m.selected >= len(m.items) {
				m.selected = 0
			}
		case key.Matches(msg, m.keys.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = len(m.items) - 1
			}
		case key.Matches(msg, m.keys.Enter):
			m.make = m.items[m.selected].make
			return m, nil, true
		case key.Matches(msg, m.keys.Back):
			m.make = 8
			return m, nil, true
		case key.Matches(msg, m.keys.Quit):
			m.make = -1
			return m, tea.Quit, false
		}
	}
	return m, nil, false
}

// View перерисовывает экран после обновления
func (m homeModel) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keys.Up,
		m.keys.Down,
		m.keys.Enter,
		m.keys.Back,
		m.keys.Quit,
	})

	if m.make == -1 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("\nGoodbye 👋\n")
	}
	title := titleStyle.Render("Добро пожаловать в хранилище GophKeeper")
	items := ""
	for i := 0; i < len(m.items); i++ {
		if i == m.selected {
			items += selectedItemStyle.Render(m.items[i].title)
		} else if i < len(m.items) {
			items += itemStyle.Render(m.items[i].title)
		}
		items += "\n"
	}
	options := fmt.Sprintf(
		"\n\n%s\n%s\n",
		items,
		errorStyle.Render(m.err),
	)

	screen := borderStyle.Render(centerWidth80Style.Render(fmt.Sprintf("%s\n\n\n%s",
		title,
		options,
	)))

	return lipgloss.JoinHorizontal(lipgloss.Top, screen) + "\n\n" + help
}
