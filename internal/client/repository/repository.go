// Package repository реализует хранение локальных данных в SQLlite базе данных
package repository

import (
	"database/sql"
	"gophkeeper/pkg/errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file" // используется для миграций
	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite" // sllite драйвер
)

// SQLLite оболочка над sql.DB
type SQLLite struct {
	*sql.DB
	Logger *log.Logger
}

// New инициализирует базу данных и осуществляет миграции
func New(logger *log.Logger) (database *SQLLite, err error) {
	database = &SQLLite{Logger: logger}
	database.DB, err = sql.Open("sqlite", "./data/gk.db?_foreign_keys=on")
	if err != nil {
		database.Logger.Debug("Error connect to database.", err)
		return nil, err
	}
	if err = database.DB.Ping(); err != nil {
		database.Logger.Debugf("cannot ping sqlite db: %s", err)
		return nil, err
	}
	if err := database.runMigrations(); err != nil {
		return nil, err
	}
	return database, nil
}

// runMigrations осуществляет миграции
func (s *SQLLite) runMigrations() error {
	driver, err := sqlite.WithInstance(s.DB, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./internal/client/migrations",
		"sqlite",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// DropMigrations осуществляет полный снос всей базы данных и заново запускает миграции
func (s *SQLLite) DropMigrations() error {
	// Получение списка всех пользовательских таблиц
	rows, err := s.DB.Query(`
        SELECT name FROM sqlite_master 
        WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name != 'schema_migrations';
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Удаление всех пользовательских таблиц
	for _, table := range tables {
		s.Logger.Debugf("Удаление таблицы: %s", table)
		if _, err := tx.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			return err
		}
	}

	// Удаление таблицы миграций
	if _, err := tx.Exec("DROP TABLE IF EXISTS schema_migrations;"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return s.runMigrations()
}

// GetLastSync получает дату последней синхронизации с сервером
func (s *SQLLite) GetLastSync() (time.Time, error) {
	row := s.QueryRow("SELECT sync_time FROM gk_sync ORDER BY sync_time DESC;")
	var t time.Time
	err := row.Scan(&t)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	return t, nil
}

// UpdateLastSync обновляет дату последней синхронизации с сервером
func (s *SQLLite) UpdateLastSync(t time.Time) error {
	_, err := s.Exec(`INSERT INTO gk_sync (sync_time) VALUES (?);`, t)
	if err != nil {
		return err
	}

	return nil
}
