// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран работы с текстовыми данными пользователя
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

// textModel хранит модель экран работы с текстовыми данными пользователя
type textModel struct {
	keys       KeyMap
	help       help.Model
	data       []domain.Text
	actionData domain.Text

	width   int
	height  int
	focus   int
	table   table.Model
	inputs  []textinput.Model
	tarea   textarea.Model
	focused int
	err     string

	showCard bool
	action   int
	index    int
}

// описывают доступ к полям редактора
const (
	tlocalid = iota
	tid
	tname
	tdata
	tdesc
)

// newTextModel инициализирует модель экран работы с текстовыми данными пользователя
func newTextModel() textModel {
	var inputs []textinput.Model = make([]textinput.Model, tdesc+1)

	inputs[tlocalid] = textinput.New()

	inputs[tid] = textinput.New()

	inputs[tdata] = textinput.New()

	inputs[tname] = textinput.New()
	inputs[tname].Placeholder = "Название данных"
	inputs[tname].CharLimit = 30
	inputs[tname].Width = 35
	inputs[tname].Prompt = ""

	inputs[tdesc] = textinput.New()
	inputs[tdesc].Placeholder = "Например: наброски сценария к будущему шедевру"
	inputs[tdesc].CharLimit = 255
	inputs[tdesc].Width = 35
	inputs[tdesc].Prompt = ""

	ta := textarea.New()
	ta.Placeholder = "Например: Любые ваши секретные данные"
	ta.Prompt = "┃ "
	ta.SetWidth(30)
	ta.SetHeight(6)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return textModel{
		keys:   DefaultKeyMap(),
		help:   help.New(),
		inputs: inputs,
		tarea:  ta,
		index:  -1,
	}
}

// Init заполняет таблицу предварительно полученными данными
func (m *textModel) Init() tea.Cmd {
	columns := []table.Column{
		{Title: "LocalID", Width: 0},
		{Title: "ID", Width: 0},
		{Title: "Название", Width: 10},
		{Title: "Комментарий", Width: 20},
		{Title: "Текст", Width: 0},
	}

	rows := make([]table.Row, len(m.data))
	for i := range m.data {
		rows[i] = table.Row{
			fmt.Sprint(m.data[i].LocalID),
			fmt.Sprint(m.data[i].ID),
			m.data[i].Name,
			m.data[i].Description,
			m.data[i].Data,
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
func (m textModel) Update(msg tea.Msg) (textModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Tab):
			if m.table.Focused() {
				m.table.Blur()
				m.focus = tname
			} else {
				m.focus++
				if m.focus > tdesc {
					m.focus = tname
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
	m.tarea, _ = m.tarea.Update(msg)
	if m.focus == tdata {
		m.tarea.Focus()
	} else {
		m.tarea.Blur()
	}

	return m, cmd
}

// focusToData переводит фокус на зону редактирования
func (m *textModel) focusToData() {
	m.showCard = true
	m.table.Blur()
	m.focus = tname
}

// focusToTable переводит фокус в таблицу выбора и скрывает зону редактирования
func (m *textModel) focusToTable() {
	m.table.Focus()
	m.showCard = false
}

// setDataToRightPanel устанавливает данные в зону редактирования
func (m *textModel) setDataToRightPanel() {
	rows := m.table.Rows()
	rows[m.table.Cursor()] = table.Row{
		fmt.Sprint(m.actionData.LocalID),
		fmt.Sprint(m.actionData.ID),
		m.actionData.Name,
		m.actionData.Data,
		m.actionData.Description,
	}
	for j := range m.inputs {
		m.inputs[j].SetValue(rows[m.table.Cursor()][j])
	}
	m.table.SetRows(rows)
	m.tarea.SetValue(m.actionData.Data)
}

// add добавляет пустую строку в таблице
func (m *textModel) add() {
	for j := range m.inputs {
		m.inputs[j].SetValue("")
	}
	m.tarea.SetValue("")
	m.inputs[tlocalid].SetValue("-1")
	m.inputs[tid].SetValue("-1")
	rows := m.table.Rows()
	rows = append(rows, make([]string, len(m.table.Columns())))
	m.table.SetRows(rows)
	m.table.SetCursor(len(rows) - 1)
	m.err = ""
}

// edit устанавливает данные в таблицу
func (m *textModel) edit() {
	rows := m.table.Rows()
	for i := range m.inputs {
		rows[m.table.Cursor()][i] = m.inputs[i].Value()
	}
	rows[m.table.Cursor()][tdata] = m.tarea.Value()
	m.table.SetRows(rows)
}

// delete удаляет строку из таблицы
func (m *textModel) delete() {
	rows := m.table.Rows()
	rows = append(rows[:m.table.Cursor()], rows[m.table.Cursor()+1:]...)
	m.table.SetRows(rows)
	if m.table.Cursor() > 0 {
		m.table.SetCursor(m.table.Cursor() - 1)
	}
}

// set устанавливает данные для передачи сервису
func (m *textModel) set() {
	row := m.table.Rows()[m.table.Cursor()]
	lid, err := strconv.Atoi(row[tlocalid])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	id, err := strconv.Atoi(row[tid])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	m.actionData = domain.Text{
		LocalID:     lid,
		ID:          id,
		Name:        row[tname],
		Data:        row[tdata],
		Description: row[tdesc],
	}
	m.tarea.SetValue(row[tdata])
}

// View перерисовывает экран после обновления
func (m textModel) View() string {
	help := m.help.FullHelpView(m.keys.fullNavHelp())
	if !m.table.Focused() {
		help = m.help.FullHelpView(m.keys.fullNavEditor())
	}

	card := ""
	if m.showCard {
		card = borderStyle.Render(fmt.Sprintf("\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n\n%s\n",
			inputStyle.Width(35).Render("Название"),
			m.inputs[tname].View(),
			inputStyle.Width(35).Render("Текст"),
			m.tarea.View(),
			inputStyle.Width(35).Render("Комментарий"),
			m.inputs[tdesc].View(),
			errorStyle.Render(m.err),
		))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, borderStyle.Render(m.table.View()), card) + "\n\n" + help
}
