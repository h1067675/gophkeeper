// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран работы с данными паролей пользователя
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

// passwordsModel хранит модель экран работы с текстовыми данными пользователя
type passwordsModel struct {
	keys       KeyMap
	help       help.Model
	data       []domain.Password
	actionData domain.Password

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
	plocalid = iota
	pid
	plogin
	ppassword
	pdomain
	pdesc
)

// newPasswordsModel инициализирует модель экран работы с данными паролей пользователя
func newPasswordsModel() *passwordsModel {
	var inputs []textinput.Model = make([]textinput.Model, pdesc+1)

	inputs[plocalid] = textinput.New()

	inputs[pid] = textinput.New()

	inputs[plogin] = textinput.New()
	inputs[plogin].Placeholder = "Логин"
	inputs[plogin].CharLimit = 255
	inputs[plogin].Width = 35
	inputs[plogin].Prompt = ""

	inputs[ppassword] = textinput.New()
	inputs[ppassword].Placeholder = "Пароль"
	inputs[ppassword].CharLimit = 255
	inputs[ppassword].Width = 35
	inputs[ppassword].Prompt = ""

	inputs[pdomain] = textinput.New()
	inputs[pdomain].Placeholder = "Сайт/приложение"
	inputs[pdomain].CharLimit = 255
	inputs[pdomain].Width = 35
	inputs[pdomain].Prompt = ""

	inputs[pdesc] = textinput.New()
	inputs[pdesc].Placeholder = "Например: изменен 01.01.2025"
	inputs[pdesc].CharLimit = 1024
	inputs[pdesc].Width = 35
	inputs[pdesc].Prompt = ""

	return &passwordsModel{
		keys:   DefaultKeyMap(),
		help:   help.New(),
		inputs: inputs,
		index:  -1,
	}
}

// Init заполняет таблицу предварительно полученными данными
func (m *passwordsModel) Init() tea.Cmd {
	columns := []table.Column{
		{Title: "LocalID", Width: 0},
		{Title: "ID", Width: 0},
		{Title: "Логин", Width: 10},
		{Title: "Пароль", Width: 0},
		{Title: "Сайт", Width: 20},
		{Title: "Комментарий", Width: 0},
	}

	rows := make([]table.Row, len(m.data))
	for i := range m.data {
		rows[i] = table.Row{
			fmt.Sprint(m.data[i].LocalID),
			fmt.Sprint(m.data[i].ID),
			m.data[i].Login,
			m.data[i].Password,
			m.data[i].Domain,
			m.data[i].Description,
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
func (m passwordsModel) Update(msg tea.Msg) (passwordsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Tab):
			if m.table.Focused() {
				m.table.Blur()
				m.focus = plogin
			} else {
				m.focus++
				if m.focus > pdesc {
					m.focus = plogin
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
				m.add()
				m.focusToData()
			}
		case key.Matches(msg, m.keys.Remove):
			if m.table.Focused() {
				m.action = delete
				m.set()
				m.delete()
				m.focusToTable()
			}
		case key.Matches(msg, m.keys.Save):
			if !m.table.Focused() {
				m.action = save
				m.edit()
				m.set()
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
				m.set()
				m.focusToData()
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

// focusToData переводит фокус на зону редактирования
func (m *passwordsModel) focusToData() {
	m.showCard = true
	m.table.Blur()
	m.focus = plogin
}

// focusToTable переводит фокус в таблицу выбора и скрывает зону редактирования
func (m *passwordsModel) focusToTable() {
	m.table.Focus()
	m.showCard = false
}

// setDataToRightPanel устанавливает данные в зону редактирования
func (m *passwordsModel) setDataToRightPanel() {
	rows := m.table.Rows()
	rows[m.table.Cursor()] = table.Row{
		fmt.Sprint(m.actionData.LocalID),
		fmt.Sprint(m.actionData.ID),
		m.actionData.Login,
		m.actionData.Password,
		m.actionData.Domain,
		m.actionData.Description,
	}
	for j := range m.inputs {
		m.inputs[j].SetValue(rows[m.table.Cursor()][j])
	}
	m.table.SetRows(rows)
}

// add добавляет пустую строку в таблице
func (m *passwordsModel) add() {
	for j := range m.inputs {
		m.inputs[j].SetValue("")
	}
	m.inputs[plocalid].SetValue("-1")
	m.inputs[pid].SetValue("-1")
	rows := m.table.Rows()
	rows = append(rows, make([]string, len(m.table.Columns())))
	m.table.SetRows(rows)
	m.table.SetCursor(len(rows) - 1)
	m.err = ""
}

// edit устанавливает данные в таблицу
func (m *passwordsModel) edit() {
	rows := m.table.Rows()
	for i := range m.inputs {
		rows[m.table.Cursor()][i] = m.inputs[i].Value()
	}
	m.table.SetRows(rows)
}

// delete удаляет строку из таблицы
func (m *passwordsModel) delete() {
	rows := m.table.Rows()
	rows = append(rows[:m.table.Cursor()], rows[m.table.Cursor()+1:]...)
	m.table.SetRows(rows)
	if m.table.Cursor() > 0 {
		m.table.SetCursor(m.table.Cursor() - 1)
	}
}

// set устанавливает данные для передачи сервису
func (m *passwordsModel) set() {
	row := m.table.Rows()[m.table.Cursor()]
	lid, err := strconv.Atoi(row[plocalid])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	id, err := strconv.Atoi(row[pid])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	m.actionData = domain.Password{
		LocalID:     lid,
		ID:          id,
		Login:       row[plogin],
		Password:    row[ppassword],
		Domain:      row[pdomain],
		Description: row[pdesc],
	}
}

// View перерисовывает экран после обновления
func (m passwordsModel) View() string {
	help := m.help.FullHelpView(m.keys.fullNavHelp())
	if !m.table.Focused() {
		help = m.help.FullHelpView(m.keys.fullNavEditor())
	}

	card := ""
	if m.showCard {
		card = borderStyle.Render(fmt.Sprintf("\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n\n\n\n%s\n",
			inputStyle.Width(35).Render("Логин"),
			m.inputs[plogin].View(),
			inputStyle.Width(35).Render("Пароль"),
			m.inputs[ppassword].View(),
			inputStyle.Width(35).Render("Сайт/приложение"),
			m.inputs[pdomain].View(),
			inputStyle.Width(35).Render("Комментарий"),
			m.inputs[pdesc].View(),
			errorStyle.Render(m.err),
		))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, borderStyle.Render(m.table.View()), card) + "\n\n" + help
}
