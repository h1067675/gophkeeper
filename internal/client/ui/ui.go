// Package ui реализует графический терминальный интерфейс пользователя основанный на BubbleTea UI
package ui

import (
	"context"
	"gophkeeper/internal/client/gophkeeper"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
)

// BubbleTeaUI описывает структуру графического интерфейса
type BubbleTeaUI struct {
	*tea.Program
	service *gophkeeper.Service
	cancel  context.CancelFunc
	Logger  *log.Logger
}

// New инициализирует новый интерфейс и делает связт с основным сервисом
func New(service *gophkeeper.Service, logger *log.Logger) *BubbleTeaUI {
	t := &BubbleTeaUI{Logger: logger, service: service}
	return t
}

// Run запускает интерфейс обеспечивая остановку основного потока приложения через cancel
func (b *BubbleTeaUI) Run(cancel context.CancelFunc) {
	b.Program = tea.NewProgram(NewModel(b.service))
	if _, err := b.Program.Run(); err != nil {
		b.Logger.Debug(err)
	}
	cancel()
}
