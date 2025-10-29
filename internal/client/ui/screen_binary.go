// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает экран работы с бинарными данными пользователя
package ui

import (
	"encoding/hex"
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

// binaryModel хранит модель экран работы с бинарными данными пользователя
type binaryModel struct {
	keys       KeyMap
	help       help.Model
	data       []domain.Binary
	actionData domain.Binary

	focus   int
	table   table.Model
	inputs  []textinput.Model
	focused int
	err     string

	showCard bool
	action   int
	index    int

	hexedit      bool
	bdata        []byte
	grid         [][]textinput.Model
	width        int
	height       int
	cursorX      int
	cursorY      int
	scrollOffset int
	viewHeight   int
}

// описывают доступ к полям редактора
const (
	bocalid = iota
	bid
	bname
	bdata
	bdesc
)

// newBinaryModel инициализирует модель экрана работы с бинарными данными пользователя
func newBinaryModel() *binaryModel {
	var inputs []textinput.Model = make([]textinput.Model, bdesc+1)

	inputs[bocalid] = textinput.New()

	inputs[bid] = textinput.New()

	inputs[bname] = textinput.New()
	inputs[bname].Placeholder = "Например: Мои бинарные данные"
	inputs[bname].CharLimit = 255
	inputs[bname].Width = 50
	inputs[bname].Prompt = ""

	inputs[bdata] = textinput.New()
	inputs[bdata].Placeholder = "Вставьте бинарные данные в текстовом виде"
	inputs[bdata].Width = 50
	inputs[bdata].Prompt = ""

	inputs[bdesc] = textinput.New()
	inputs[bdesc].Placeholder = "Тут должен быть комментарий к данным"
	inputs[bdesc].CharLimit = 255
	inputs[bdesc].Width = 50
	inputs[bdesc].Prompt = ""

	return &binaryModel{
		keys:       DefaultKeyMap(),
		help:       help.New(),
		inputs:     inputs,
		index:      -1,
		viewHeight: 8,
		width:      6,
	}
}

// Init заполняет таблицу предварительно полученными данными
func (m *binaryModel) Init() tea.Cmd {
	columns := []table.Column{
		{Title: "LocalID", Width: 0},
		{Title: "ID", Width: 0},
		{Title: "Название", Width: 10},
		{Title: "Текст", Width: 0},
		{Title: "Комментарий", Width: 20},
	}

	rows := make([]table.Row, len(m.data))
	for i := range m.data {
		rows[i] = table.Row{
			fmt.Sprint(m.data[i].LocalID),
			fmt.Sprint(m.data[i].ID),
			m.data[i].Name,
			m.data[i].Data,
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
func (m binaryModel) Update(msg tea.Msg) (binaryModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case m.table.Focused():
			switch {
			case key.Matches(msg, m.keys.Tab):
				m.table.Blur()
				m.focus = bname
			case key.Matches(msg, m.keys.Up):
				m.table.MoveUp(1)
			case key.Matches(msg, m.keys.Down):
				m.table.MoveDown(1)
			case key.Matches(msg, m.keys.Add):
				m.add()
				m.focusToData()
			case key.Matches(msg, m.keys.Remove):
				m.action = delete
				m.set()
				m.delete()
				m.focusToTable()
			case key.Matches(msg, m.keys.Back):
				m.action = back
			case key.Matches(msg, m.keys.Enter):
				m.action = load
				m.set()
				m.focusToData()
			}
		default:
			switch {
			case key.Matches(msg, m.keys.Save):
				m.action = save
				m.edit()
				m.set()
				m.focusToTable()
			case key.Matches(msg, m.keys.Back):
				m.action = none
				m.focusToTable()
			case key.Matches(msg, m.keys.Tab):
				m.focus++
				m.hexedit = false
				if len(m.grid) > 0 {
					m.cursorY, m.cursorX = 0, 0
					m.grid[m.cursorY][m.cursorX].Blur()
				}
				if m.focus == tdesc+1 && len(m.grid) > 0 {
					m.hexedit = true
					m.grid[0][0].Focus()
				} else if m.focus >= tdesc+1 {
					m.focus = bname
				}
			case m.hexedit:
				switch {
				case key.Matches(msg, m.keys.Right):
					m.grid[m.cursorY][m.cursorX].Blur()
					if m.cursorX < m.width-1 {
						m.cursorX++
					}
					if m.cursorY*m.width+m.cursorX > len(m.bdata) {
						m.cursorY, m.cursorX = len(m.bdata)/m.width, len(m.bdata)%m.width
					}
					m.grid[m.cursorY][m.cursorX].Focus()
				case key.Matches(msg, m.keys.Left):
					m.grid[m.cursorY][m.cursorX].Blur()
					if m.cursorX > 0 {
						m.cursorX--
					}
					m.grid[m.cursorY][m.cursorX].Focus()
				case key.Matches(msg, m.keys.Down):
					m.grid[m.cursorY][m.cursorX].Blur()
					if m.cursorY < len(m.grid)-1 {
						m.cursorY++
						if m.cursorY*m.width+m.cursorX > len(m.bdata) {
							m.cursorY, m.cursorX = len(m.bdata)/m.width, len(m.bdata)%m.width
						}
						if m.cursorY >= m.scrollOffset+m.viewHeight {
							m.scrollOffset++
						}
					} else if m.cursorY == len(m.bdata)/m.width-1 && len(m.bdata)%m.width == 0 {
						m.cursorY, m.cursorX = len(m.bdata)/m.width, len(m.bdata)%m.width
						m.addRowHEXEditor()
					}
					m.grid[m.cursorY][m.cursorX].Focus()
				case key.Matches(msg, m.keys.Up):
					m.grid[m.cursorY][m.cursorX].Blur()
					if m.cursorY > 0 {
						m.cursorY--
						if m.cursorY < m.scrollOffset {
							m.scrollOffset--
							if m.scrollOffset < 0 {
								m.scrollOffset = 0
							}
						}
					}
					m.grid[m.cursorY][m.cursorX].Focus()
				}
			default:

			}
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

	if !m.table.Focused() {
		m.changeBinaryData()
	}

	if len(m.grid) > 0 && m.cursorY < len(m.grid) {
		m.grid[m.cursorY][m.cursorX], cmd = m.grid[m.cursorY][m.cursorX].Update(msg)
	}
	return m, cmd
}

// focusToData переводит фокус на зону редактирования
func (m *binaryModel) focusToData() {
	m.showCard = true
	m.table.Blur()
	m.focus = bname
}

// focusToTable переводит фокус в таблицу выбора и скрывает зону редактирования
func (m *binaryModel) focusToTable() {
	m.table.Focus()
	m.showCard = false
}

// setDataToRightPanel устанавливает данные в зону редактирования
func (m *binaryModel) setDataToRightPanel() {
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
	m.fillHEXEditor()
}

// add добавляет пустую строку в таблице
func (m *binaryModel) add() {
	for j := range m.inputs {
		m.inputs[j].SetValue("")
	}
	m.grid = [][]textinput.Model{}
	m.inputs[bocalid].SetValue("-1")
	m.inputs[bid].SetValue("-1")
	rows := m.table.Rows()
	rows = append(rows, make([]string, len(m.table.Columns())))
	m.table.SetRows(rows)
	m.table.SetCursor(len(rows) - 1)
	m.err = ""
}

// edit устанавливает данные в таблицу
func (m *binaryModel) edit() {
	rows := m.table.Rows()
	for i := range m.inputs {
		rows[m.table.Cursor()][i] = m.inputs[i].Value()
	}
	rows[m.table.Cursor()][tdata] = string(m.bdata)
	m.table.SetRows(rows)
}

// delete удаляет строку из таблицы
func (m *binaryModel) delete() {
	rows := m.table.Rows()
	rows = append(rows[:m.table.Cursor()], rows[m.table.Cursor()+1:]...)
	m.table.SetRows(rows)
	if m.table.Cursor() > 0 {
		m.table.SetCursor(m.table.Cursor() - 1)
	}
}

// set устанавливает данные для передачи сервису
func (m *binaryModel) set() {
	row := m.table.Rows()[m.table.Cursor()]
	lid, err := strconv.Atoi(row[bocalid])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	id, err := strconv.Atoi(row[bid])
	if err != nil {
		m.err = "Ошибка чтения данных"
	}
	m.actionData = domain.Binary{
		LocalID:     lid,
		ID:          id,
		Name:        row[bname],
		Data:        row[bdata],
		Description: row[bdesc],
	}
	m.fillHEXEditor()
}

// fillHEXEditor заполняет HEX редактор данными из строкового представления
func (m *binaryModel) fillHEXEditor() {
	rs := (len(m.bdata) + m.width - 1) / m.width
	m.grid = make([][]textinput.Model, rs)
	for r := range m.grid {
		m.grid[r] = make([]textinput.Model, m.width)
		for c := 0; c < m.width; c++ {
			i := r*m.width + c
			ti := textinput.New()
			ti.CharLimit = 2
			ti.Width = 2
			ti.Prompt = ""
			ti.SetValue("")
			if i < len(m.bdata) {
				ti.SetValue(fmt.Sprintf("%02X", m.bdata[i]))
			}
			m.grid[r][c] = ti
		}
	}
	m.cursorY, m.cursorX = 0, 0
}

// addRowHEXEditor добавляет строку в HEX редактор
func (m *binaryModel) addRowHEXEditor() {
	r := len(m.grid)
	m.grid = append(m.grid, make([]textinput.Model, m.width))
	for c := 0; c < m.width; c++ {
		ti := textinput.New()
		ti.CharLimit = 2
		ti.Width = 2
		ti.Prompt = ""
		ti.SetValue("")
		m.grid[r][c] = ti
	}
}

// changeBinaryData при изменении даннх либо в строке либо в HEX редакторе меняет данные там где они не менялись
func (m *binaryModel) changeBinaryData() {
	if m.focus != bdesc+1 && m.inputs[bdata].Value() != string(m.bdata) {
		m.bdata = []byte(m.inputs[bdata].Value())
		m.fillHEXEditor()
	} else if m.hexedit && len(m.grid) > 0 {
		val := m.grid[m.cursorY][m.cursorX].Value()
		if len(val) == 1 {
			val = "0" + val
		}
		b, err := hex.DecodeString(val)
		if err == nil && len(b) == 1 {
			i := m.cursorY*m.width + m.cursorX
			if i < len(m.bdata) {
				m.bdata[i] = b[0]
			} else {
				m.bdata = append(m.bdata, b[0])
			}
		}
		m.inputs[bdata].SetValue(string(m.bdata))
	}
}

// View перерисовывает экран после обновления
func (m binaryModel) View() string {
	help := m.help.FullHelpView(m.keys.fullNavHelp())
	if !m.table.Focused() {
		help = m.help.FullHelpView(m.keys.fullNavHEX())
	}

	s := ""
	end := m.scrollOffset + m.viewHeight
	if end > len(m.grid) {
		end = len(m.grid)
	}
	for r := m.scrollOffset; r < end; r++ {
		s += fmt.Sprintf("%08X  ", r*m.width)
		for c := range m.grid[r] {
			s += fmt.Sprintf("%s ", m.grid[r][c].View())
		}

		s += " |"
		for c := range m.grid[r] {
			i := r*m.width + c
			if i < len(m.bdata) {
				b := m.bdata[i]
				if b >= 32 && b <= 126 {
					s += string(b)
				} else {
					s += "."
				}
			}
		}
		s += "|\n"
	}

	card := ""
	if m.showCard {
		card = borderStyle.Render(fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n\n%s\n\n%s",
			inputStyle.Width(35).Render("Название"),
			m.inputs[bname].View(),
			inputStyle.Width(35).Render("Бинарная строка в текстовом формате"),
			m.inputs[bdata].View(),
			inputStyle.Width(35).Render("Комментарий"),
			m.inputs[bdesc].View(),
			s,
			errorStyle.Render(m.err),
		))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, borderStyle.Render(m.table.View()), card) + "\n\n" + help
}
