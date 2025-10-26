// Package domain описывает форматы сущностей всего проекта
package domain

import "time"

type (
	// Password сущность для работы с данными паролей пользователя
	Password struct {
		ID          int       `json:"id" db:"id"`
		LocalID     int       `json:"local_id" db:"-"`
		UserID      int       `json:"user_id" db:"users_id"`
		Login       string    `json:"login" db:"login"`
		Password    string    `json:"password" db:"password"`
		Domain      string    `json:"domain" db:"domain"`
		CreatedAt   time.Time `json:"created_at" db:"created_at"`
		ChangedAt   time.Time `json:"changed_at" db:"changed_at"`
		Description string    `json:"description" db:"description"`
		Deleted     bool      `json:"deleted" db:"deleted"`
	}

	// Card сущность для работы с данными банковских карт пользователя
	Card struct {
		ID          int       `json:"id" db:"id"`
		LocalID     int       `json:"local_id" db:"-"`
		UserID      int       `json:"user_id" db:"users_id"`
		Bank        string    `json:"bank" db:"bank"`
		Number      string    `json:"number" db:"number"`
		Month       string    `json:"month" db:"month"`
		Year        string    `json:"year" db:"year"`
		CVC         string    `json:"cvc" db:"cvc"`
		Holder      string    `json:"holder" db:"holder"`
		CreatedAt   time.Time `json:"created_at" db:"created_at"`
		ChangedAt   time.Time `json:"changed_at" db:"changed_at"`
		Description string    `json:"description" db:"description"`
		Deleted     bool      `json:"deleted" db:"deleted"`
	}

	// Text сущность для работы с текстовыми данными пользователя
	Text struct {
		ID          int       `json:"id" db:"id"`
		LocalID     int       `json:"local_id" db:"-"`
		UserID      int       `json:"user_id" db:"users_id"`
		Name        string    `json:"name" db:"name"`
		Data        string    `json:"data" db:"data"`
		CreatedAt   time.Time `json:"created_at" db:"created_at"`
		ChangedAt   time.Time `json:"changed_at" db:"changed_at"`
		Description string    `json:"description" db:"description"`
		Deleted     bool      `json:"deleted" db:"deleted"`
	}

	// Binary сущность для работы с бинарными данными пользователя
	Binary struct {
		ID          int       `json:"id" db:"id"`
		LocalID     int       `json:"local_id" db:"-"`
		UserID      int       `json:"user_id" db:"users_id"`
		Name        string    `json:"name" db:"name"`
		Data        string    `json:"data" db:"data"`
		CreatedAt   time.Time `json:"created_at" db:"created_at"`
		ChangedAt   time.Time `json:"changed_at" db:"changed_at"`
		Description string    `json:"description" db:"description"`
		Deleted     bool      `json:"deleted" db:"deleted"`
	}

	// SyncPayload сущность описывает формат синхронизации сервера и клиента
	SyncPayload struct {
		Passwords    []Password `json:"passwords"`
		Cards        []Card     `json:"cards"`
		Texts        []Text     `json:"texts"`
		Binaries     []Binary   `json:"binaries"`
		AllowedToken string     `json:"allowed_token"`
		Error        string     `json:"error"`
		LastSync     time.Time  `json:"last_sync"`
	}
)

// TableName возвращает название таблицы
func (Password) TableName() string { return "gk_passwords" }

// TableName возвращает название таблицы
func (Card) TableName() string { return "gk_bank_cards" }

// TableName возвращает название таблицы
func (Text) TableName() string { return "gk_text_data" }

// TableName возвращает название таблицы
func (Binary) TableName() string { return "gk_binary_data" }
