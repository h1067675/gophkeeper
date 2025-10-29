// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран аторизации по одноразовому паролю
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// totpModel хранит модель экрана двухфакторной авторизации
type totpModel struct {
	keys KeyMap
	help help.Model

	items        []item
	code         textinput.Model
	email        textinput.Model
	qr           string
	recoveryStep int
	err          string
	focus        int
	done         bool
	selected     int
	make         int
	showQR       bool
}

// newTOTPModel инициализирует модель экрана двухфакторной авторизации
func newTOTPModel() *totpModel {
	code := textinput.New()
	code.Placeholder = "Код подтверждения"
	code.Focus()
	code.CharLimit = 6
	code.Width = 30

	email := textinput.New()
	email.Placeholder = "Электронная почта"
	email.CharLimit = 32
	email.Width = 30

	ims := make([]item, 0)
	ims = append(ims, item{title: "Подтвердить и войти", make: 1, show: true})
	ims = append(ims, item{title: "Перейти к повторной регистрации в Google Authenticator", make: 2, show: true})
	ims = append(ims, item{title: "Отправить код подтверждения на электронную почту", make: 3, show: false})
	ims = append(ims, item{title: "Подтвердить Email код", make: 4, show: false})
	ims = append(ims, item{title: "Показать QR-код Google Authenticator", make: 5, show: false})
	ims = append(ims, item{title: "Выйти", make: -1, show: true})

	return &totpModel{
		code:  code,
		email: email,
		keys:  DefaultKeyMap(),
		help:  help.New(),
		items: ims,
	}
}

// Update обновляет данные полей и реализует основную логику экрана
func (m totpModel) Update(msg tea.Msg) (totpModel, tea.Cmd, bool) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showQR {
			m.showQR = false
			return m, nil, false
		}
		switch {
		case key.Matches(msg, m.keys.Down) || key.Matches(msg, m.keys.Tab):
			m.selected++
			for m.selected-1 < len(m.items) && !m.items[m.selected-1].show {
				m.selected++
			}
			if m.selected > len(m.items) {
				m.selected = 0
				if m.make == 2 {
					m.email.Focus()
				} else {
					m.code.Focus()
				}
			} else {
				m.code.Blur()
				m.email.Blur()
			}
			return m, nil, false
		case key.Matches(msg, m.keys.Up):
			m.selected--
			for m.selected-1 >= 0 && !m.items[m.selected-1].show {
				m.selected--
			}
			if m.selected < 0 {
				m.selected = len(m.items)
				m.code.Blur()
				m.email.Blur()
			} else if m.selected == 0 {
				if m.make == 2 {
					m.email.Focus()
				} else {
					m.code.Focus()
				}
			}
			return m, nil, false
		case key.Matches(msg, m.keys.Enter):
			if m.selected > 0 {
				m.make = m.items[m.selected-1].make
			}
			switch m.make {
			case 1, 4:
				if m.code.Value() == "" {
					m.err = "Ошибка! Код подтверждения не может быть пустым"
					return m, nil, false
				}
				m.items[0].show = true
				m.items[1].show = true
				m.items[2].show = false
				m.items[3].show = false
				m.items[4].show = false
				return m, nil, true
			case 2:
				m.items[0].show = false
				m.items[1].show = false
				m.items[2].show = true
				m.items[3].show = false
				m.items[4].show = false
			case 3:
				if m.email.Value() == "" {
					m.err = "Ошибка! Email не может быть пустым"
					return m, nil, false
				}
				m.items[0].show = false
				m.items[1].show = false
				m.items[2].show = false
				m.items[3].show = true
				m.items[4].show = false
				return m, nil, true
			case 5:
				return m, nil, true
			}
			return m, cmd, false
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit, false
		case key.Matches(msg, m.keys.Back):
			m.make = 6
			return m, nil, true
		case key.Matches(msg, m.keys.QR):
			m.make = 5
			return m, nil, true
		}
	}

	m.code, _ = m.code.Update(msg)
	m.email, _ = m.email.Update(msg)
	return m, cmd, false
}

// View перерисовывает экран после обновления
func (m totpModel) View() string {

	help := m.help.ShortHelpView([]key.Binding{
		m.keys.Tab,
		m.keys.Enter,
		m.keys.Back,
		m.keys.Quit,
	})
	if m.qr != "" {
		help = m.help.ShortHelpView([]key.Binding{
			m.keys.Tab,
			m.keys.QR,
			m.keys.Enter,
			m.keys.Back,
			m.keys.Quit,
		})
		m.items[4].show = true
	}

	title := titleStyle.Render("Введите код подтверждения из Google Authenticator")
	input := m.code.View()
	items := ""
	for i := 0; i < len(m.items); i++ {
		if i == m.selected-1 && m.items[i].show {
			items += selectedItemStyle.Render(m.items[i].title) + "\n"
		} else if i < len(m.items) {
			if m.items[i].show {
				items += itemStyle.Render(m.items[i].title) + "\n"
			}
		}
	}
	options := fmt.Sprintf(
		"\n\n%s\n\n",
		items,
	)

	switch m.make {
	case 2:
		title = titleStyle.Render("Введите вашу электронную почту указанную при регистрации для \n повторной регистрации одноразовых паролей в Google Authenticator")
		input = m.email.View()
	case 3:
		title = titleStyle.Render("Введите код подтверждения отправленный вам на электронную почту")
		input = m.code.View()
	}

	if m.showQR {
		return m.qr + "\n Для выхода из режима просмотра QR-кода нажмите любую кнопку"
	}

	screen := borderStyle.Render(centerWidth80Style.Render(fmt.Sprintf("%s\n\n%s\n%s\n%s\n",
		title,
		input,
		options,
		errorStyle.Render(m.err),
	)))

	return lipgloss.JoinHorizontal(lipgloss.Top, screen) + "\n\n" + help
}
