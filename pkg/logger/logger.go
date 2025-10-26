// Package logger осуществляет настройку и ведение лога
package logger

import (
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// NewLogFile создает новый файл логгов
func NewLogFile(filename string) (*os.File, error) {
	// получаем директорию текущего файла
	appfile, err := os.Executable()
	if err != nil {
		return nil, err
	}
	// читаем файл конфигурации
	logFile, err := os.Create(filepath.Join(filepath.Dir(appfile), filename))
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

// InitializeLogger инициализирует и настраивает логгер
func InitializeLogger(format log.Formatter, level log.Level, output io.Writer) *log.Logger {
	var logger = log.New()
	logger.SetFormatter(format)
	logger.SetOutput(output)
	logger.SetLevel(level)
	return logger
}
