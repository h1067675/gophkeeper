// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
//
// описывает клавиши используемые в приложении
package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap — общая карта клавиш, доступная всем экранам
type KeyMap struct {
	Tab      key.Binding
	Prev     key.Binding
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Action   key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Edit     key.Binding
	Save     key.Binding
	Add      key.Binding
	Remove   key.Binding
	Register key.Binding
	QR       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

// DefaultKeyMap — создаём общие биндинги
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "переключить"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("Shift+Tab", "предыдущая"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("⇧", "вверх"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("⇩", "вниз"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("⇦", "влево"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("⇨", "вправо"),
		),
		Action: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "действие"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "подтвердить"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "назад"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("q", "выйти"),
		),
		Register: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("Ctrl+r", "сохранить"),
		),
		Edit: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("Ctrl+e", "редактировать"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("Ctrl+s", "сохранить"),
		),
		Add: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("Ctrl+n", "добавить"),
		),
		Remove: key.NewBinding(
			key.WithKeys("ctrl+w"),
			key.WithHelp("Ctrl+w", "удалить"),
		),
		QR: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("Ctrl+g", "показать QR-код"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("PgUp", "страница вверх"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdn"),
			key.WithHelp("PgDn", "страница вниз"),
		),
	}
}

// ShortHelp создает короткое описание клавиш
func (k KeyMap) ShortHelp() []key.Binding {
	keys := make([]key.Binding, 0)
	keys = append(keys, k.Tab, k.Enter, k.Back)
	return keys
}

// FullHelp создает длинное описание клавиш
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up, k.Down,
		},
		{
			k.Action, k.Back,
		},
		{
			k.Enter, k.Quit,
		},
	}
}

// fullNavHelp создает длинное описание клавиш для редакторов
func (k KeyMap) fullNavHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Tab,
			k.Back,
		},
		{
			k.Up,
			k.Down,
		},
		{
			k.Add,
			k.Remove,
		},
		{
			k.Enter,
			k.Quit,
		},
	}
}

// fullNavHEX создает длинное описание клавиш для HEX редактора
func (k KeyMap) fullNavHEX() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Tab,
			k.Back,
		},
		{
			k.Up,
			k.Down,
			k.Left,
			k.Right,
		},
		{
			k.Save,
		},
		{
			k.Quit,
		},
	}
}

// fullNavEditor создает длинное короткое клавиш для редакторов
func (k KeyMap) fullNavEditor() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Tab,
			k.Back,
		},
		{},
		{
			k.Save,
		},
		{
			k.Quit,
		},
	}
}
