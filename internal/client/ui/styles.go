// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает стили используемые в графическом интерфейсе
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// переменные описывающие графические стили
var (
	inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7"))

	titleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(lipgloss.Color("#FFD166")).
			Bold(true).
			Padding(1, 2)

	itemStyle = lipgloss.NewStyle()

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#790086ff"))

	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8DADC"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c90000ff"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#44475A")).
			Padding(1, 2).
			Height(20)

	centerWidth80Style = lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center).Width(80)

	versionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#272727ff"))
)

// item описывает элемент меню
type item struct {
	title string
	make  int
	show  bool
}

// описывают общие операции
const (
	none = iota
	save
	delete
	load
	back
)
