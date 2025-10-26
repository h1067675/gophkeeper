// Package repository реализует хранение данных сервера в PostgeSQL базе данных
package repository

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	log "github.com/sirupsen/logrus"
)

type (
	// PGXDB оболочка над sql.DB
	PGXDB struct {
		*sql.DB
		Logger *log.Logger
	}

	// Tabler интерфейсный тип необходимый для обработки запросов к базе даннх через одну функцию
	Tabler interface {
		TableName() string
	}
)

// New инициализирует базу данных и осуществляет миграции
func New(dbPath string, logger *log.Logger, drop bool) (database *PGXDB, err error) {
	database = &PGXDB{Logger: logger}
	database.DB, err = sql.Open("pgx", dbPath)
	if err != nil {
		database.Logger.Debug("Error connect to database.", err)
		return nil, err
	}
	if drop {
		err = database.downMigrations()
		if err != nil {
			return nil, err
		}
	}
	err = database.upMigrations()
	if err != nil {
		return nil, err
	}
	return database, err
}

// downMigrations осуществляет DOWN миграции функция для отладки при разработке, не применять на рабочей базе данных
func (p *PGXDB) downMigrations() error {
	driver, err := postgres.WithInstance(p.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./internal/server/migrations",
		"postgres", driver,
	)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// upMigrations осуществляет UP миграции
func (p *PGXDB) upMigrations() error {
	driver, err := postgres.WithInstance(p.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./internal/server/migrations",
		"postgres", driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (d *PGXDB) CloseRepository() error {
	return d.Close()
}
