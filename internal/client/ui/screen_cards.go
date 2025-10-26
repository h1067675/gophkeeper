// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран работы с данными банковских карт пользователя
package ui

import (
	"fmt"
	"gophkeeper/pkg/domain"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// cardsModel хранит модель экран работы с данными банковских карт пользователя
type cardsModel struct {
	keys       KeyMap
	help       help.Model
	cards      []domain.Card
	actionCard domain.Card

	width   int
	height  int
	focus   int
	table   table.Model
	inputs  []textinput.Model
	focused int
	err     string

	showCard bool
	action   int
	index    int
}

// описывают доступ к полям редактора
const (
	localid = iota
	id
	bank
	number
	month
	year
	cvc
	holder
	desc
)

// newCardsModel инициализирует модель экран работы с данными банковских карт пользователя
func newCardsModel() cardsModel {
	var inputs []textinput.Model = make([]textinput.Model, desc+1)
	inputs[localid] = textinput.New()

	inputs[id] = textinput.New()

	inputs[bank] = textinput.New()
	inputs[bank].Placeholder = "Например: Сбер - основная"
	inputs[bank].CharLimit = 30
	inputs[bank].Width = 35
	inputs[bank].Prompt = ""

	inputs[number] = textinput.New()
	inputs[number].Placeholder = "4505 **** **** 1234"
	inputs[number].CharLimit = 20
	inputs[number].Width = 35
	inputs[number].Prompt = ""

	inputs[month] = textinput.New()
	inputs[month].Placeholder = "MM "
	inputs[month].CharLimit = 2
	inputs[month].Width = 2
	inputs[month].Prompt = ""

	inputs[year] = textinput.New()
	inputs[year].Placeholder = "YY "
	inputs[year].CharLimit = 2
	inputs[year].Width = 2
	inputs[year].Prompt = ""

	inputs[cvc] = textinput.New()
	inputs[cvc].Placeholder = "XXX"
	inputs[cvc].CharLimit = 3
	inputs[cvc].Width = 7
	inputs[cvc].Prompt = ""

	inputs[holder] = textinput.New()
	inputs[holder].Placeholder = "VLADIMIR PUTIN"
	inputs[holder].CharLimit = 27
	inputs[holder].Width = 35
	inputs[holder].Prompt = ""

	inputs[desc] = textinput.New()
	inputs[desc].Placeholder = "Например: Кэшбэк 5% на рестораны"
	inputs[desc].CharLimit = 255
	inputs[desc].Width = 35
	inputs[desc].Prompt = ""

	return cardsModel{
		keys:   DefaultKeyMap(),
		help:   help.New(),
		inputs: inputs,
		index:  -1,
	}
}

// Init заполняет таблицу предварительно полученными данными
func (m *cardsModel) Init() tea.Cmd {
	columns := []table.Column{
		{Title: "LocalID", Width: 0},
		{Title: "ID", Width: 0},
		{Title: "Название", Width: 10},
		{Title: "Номер карты", Width: 20},
		{Title: "Месяц", Width: 0},
		{Title: "Год", Width: 0},
		{Title: "CVC", Width: 0},
		{Title: "Держатель", Width: 0},
		{Title: "Комментарий", Width: 0},
	}

	rows := make([]table.Row, len(m.cards))
	for i := range m.cards {
		rows[i] = table.Row{
			fmt.Sprint(m.cards[i].LocalID),
			fmt.Sprint(m.cards[i].ID),
			m.cards[i].Bank,
			m.cards[i].Number,
			m.cards[i].Month,
			m.cards[i].Year,
			m.cards[i].CVC,
			m.cards[i].Holder,
			m.cards[i].Description,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(17),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	m.table = t
	return textarea.Blink
}

// Update обновляет данные полей и реализует основную логику экрана
func (m cardsModel) Update(msg tea.Msg) (cardsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Tab):
			if m.table.Focused() {
				m.table.Blur()
				m.focus = bank
			} else {
				m.focus++
				if m.focus > desc {
					m.focus = bank
				}
			}
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up) && m.table.Focused():
			m.table.MoveUp(1)
		case key.Matches(msg, m.keys.Down) && m.table.Focused():
			m.table.MoveDown(1)
		case key.Matches(msg, m.keys.Add):
			if m.table.Focused() {
				m.addCard()
				m.focusToCard()
			}
		case key.Matches(msg, m.keys.Remove):
			if m.table.Focused() {
				m.action = delete
				m.setDataCard()
				m.deleteCard()
				m.focusToTable()
			}
		case key.Matches(msg, m.keys.Save):
			if !m.table.Focused() {
				m.action = save
				m.editCard()
				m.setDataCard()
				m.focusToTable()
			}
		case key.Matches(msg, m.keys.Back):
			if m.table.Focused() {
				m.action = back
			} else {
				m.action = none
				m.focusToTable()
			}
		case key.Matches(msg, m.keys.Enter):
			if m.table.Focused() {
				m.action = load
				m.setDataCard()
				m.focusToCard()
			}
		case key.Matches(msg, m.keys.Back):
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	for i := range m.inputs {
		m.inputs[i], _ = m.inputs[i].Update(msg)
		m.inputs[i].Blur()
		if i == m.focus {
			m.inputs[i].Focus()
		}
	}

	return m, cmd
}

// focusToCard переводит фокус на зону редактирования
func (m *cardsModel) focusToCard() {
	m.showCard = true
	m.table.Blur()
	m.focus = bank
}

// focusToTable переводит фокус в таблицу выбора и скрывает зону редактирования
func (m *cardsModel) focusToTable() {
	m.table.Focus()
	m.showCard = false
}

// setDataToCard устанавливает данные в зону редактирования
func (m *cardsModel) setDataToCard() {
	rows := m.table.Rows()
	rows[m.table.Cursor()] = table.Row{
		fmt.Sprint(m.actionCard.LocalID),
		fmt.Sprint(m.actionCard.ID),
		m.actionCard.Bank,
		m.actionCard.Number,
		m.actionCard.Month,
		m.actionCard.Year,
		m.actionCard.CVC,
		m.actionCard.Holder,
		m.actionCard.Description,
	}
	for j := range m.inputs {
		m.inputs[j].SetValue(rows[m.table.Cursor()][j])
	}
	m.table.SetRows(rows)
}

// addCard добавляет пустую строку в таблице
func (m *cardsModel) addCard() {
	for j := range m.inputs {
		m.inputs[j].SetValue("")
	}
	m.inputs[localid].SetValue("-1")
	m.inputs[id].SetValue("-1")
	rows := m.table.Rows()
	rows = append(rows, make([]string, len(m.table.Columns())))
	m.table.SetRows(rows)
	m.table.SetCursor(len(rows) - 1)
	m.err = ""
}

// editCard устанавливает данные в таблицу
func (m *cardsModel) editCard() {
	rows := m.table.Rows()
	for i := range m.inputs {
		rows[m.table.Cursor()][i] = m.inputs[i].Value()
	}
	m.table.SetRows(rows)
}

// deleteCard удаляет строку из таблицы
func (m *cardsModel) deleteCard() {
	rows := m.table.Rows()
	rows = append(rows[:m.table.Cursor()], rows[m.table.Cursor()+1:]...)
	m.table.SetRows(rows)
	if m.table.Cursor() > 0 {
		m.table.SetCursor(m.table.Cursor() - 1)
	}
}

// setDataCard устанавливает данные для передачи сервису
func (m *cardsModel) setDataCard() {
	row := m.table.Rows()[m.table.Cursor()]
	lid, err := strconv.Atoi(row[localid])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	id, err := strconv.Atoi(row[id])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	m.actionCard = domain.Card{
		LocalID:     lid,
		ID:          id,
		Bank:        row[bank],
		Number:      row[number],
		Month:       row[month],
		Year:        row[year],
		CVC:         row[cvc],
		Holder:      row[holder],
		Description: row[desc],
	}
}

// View перерисовывает экран после обновления
func (m cardsModel) View() string {
	help := m.help.FullHelpView(m.keys.fullNavHelp())
	if !m.table.Focused() {
		help = m.help.FullHelpView(m.keys.fullNavEditor())
	}

	card := ""
	if m.showCard {
		card = borderStyle.Render(fmt.Sprintf("\n%s\n%s\n\n%s\n%s\n\n%s   %s\n%s/%s            %s\n\n%s\n%s\n\n%s\n%s\n\n%s\n",
			inputStyle.Width(35).Render("Название"),
			m.inputs[bank].View(),
			inputStyle.Width(35).Render("Номер карты"),
			m.inputs[number].View(),
			inputStyle.Width(13).Render("Срок действия"),
			inputStyle.Width(7).Render("CVC код"),
			m.inputs[month].View(),
			m.inputs[year].View(),
			m.inputs[cvc].View(),
			inputStyle.Width(35).Render("Держатель"),
			m.inputs[holder].View(),
			inputStyle.Width(35).Render("Комментарий"),
			m.inputs[desc].View(),
			errorStyle.Render(m.err),
		))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, borderStyle.Render(m.table.View()), card) + "\n\n" + help
}
