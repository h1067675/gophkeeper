// Package repository реализует хранение данных сервера в PostgeSQL базе данных
package repository

import (
	"database/sql"
	"gophkeeper/pkg/domain"
	"gophkeeper/pkg/errors"
	"time"
)

// UserGetByLogin получает id пользователя по логину
func (p *PGXDB) UserGetByLogin(login string) (int, error) {
	rows := p.DB.QueryRow("SELECT id FROM gk_users WHERE login = $1;", login)
	id := -1
	err := rows.Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return -1, err
	}
	return id, nil
}

// UserGetByEmail получает id пользователя по электронной почте
func (p *PGXDB) UserGetByEmail(email string) (int, error) {
	rows := p.DB.QueryRow("SELECT id FROM gk_users WHERE email = $1;", email)
	id := -1
	err := rows.Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return -1, err
	}
	return id, nil
}

// UserGetByToken получает id пользователя уникальному токену
func (p *PGXDB) UserGetByToken(token string) (int, time.Time, error) {
	rows := p.DB.QueryRow("SELECT users_id, expire FROM gk_sessions WHERE session = $1;", token)
	id := -1
	var t time.Time
	err := rows.Scan(&id, &t)
	if err != nil {
		return -1, time.Time{}, err
	}
	return id, t, nil
}

// SessionSave сохраняет данные сессии пользователя
func (p *PGXDB) SessionSave(userid int, session string, expire time.Time, allowed bool) error {
	_, err := p.DB.Exec("INSERT INTO gk_sessions (users_id, session, expire, allowed_totp) VALUES ($1, $2, $3, $4);",
		userid, session, expire.UTC(), allowed)
	return err
}

// SessionSaveTOTPSecret сохраняет данные двухфакторной авторизации пользователя
func (p *PGXDB) SessionSaveTOTPSecret(session, secretKey string) error {
	_, err := p.DB.Exec("UPDATE gk_sessions SET key_totp = $1 WHERE session = $2", secretKey, session)
	return err
}

// SessionGetTOTPSecret получает данные двухфакторной авторизации пользователя
func (p *PGXDB) SessionGetTOTPSecret(session string) (int, string, error) {
	rows := p.DB.QueryRow("SELECT users_id, key_totp FROM gk_sessions WHERE session = $1;", session)
	id := -1
	var secretKey string
	err := rows.Scan(&id, &secretKey)
	if err != nil {
		return -1, "", err
	}
	return id, secretKey, nil
}

// SessionGet получает данные сессии пользователя
func (p *PGXDB) SessionGet(userID int) (session, secretKey string, expire time.Time, err error) {
	rows := p.DB.QueryRow("SELECT session, key_totp, expire FROM gk_sessions WHERE users_id = $1 AND allowed_totp = FALSE;", userID)
	err = rows.Scan(&session, &secretKey, &expire)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", time.Time{}, err
	}
	return session, secretKey, expire, nil
}

// SessionGetEmail получает электронный адрес пользователя из сессии для сверки и отправки на нее кода подтверждения
func (p *PGXDB) SessionGetEmail(session string) (email, code string, expireCode time.Time, err error) {
	rows := p.DB.QueryRow(`	SELECT u.email, COALESCE(s.code_email, '') , COALESCE(s.expire_code_email, current_timestamp) 
							FROM gk_sessions s
							JOIN gk_users u ON u.id = s.users_id
							WHERE s.session = $1`, session)
	err = rows.Scan(&email, &code, &expireCode)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", time.Time{}, err
	}
	return
}

// SessionSetEmailCode устанавливает код подтверждения электронного адреса пользователя для сессии
func (p *PGXDB) SessionSetEmailCode(session, code string, expireCode time.Time) (err error) {
	_, err = p.DB.Exec("UPDATE gk_sessions SET code_email = $1, expire_code_email = $2 WHERE session = $3", code, expireCode, session)
	return err
}

// UserRegistration добавляет пользователя в базу данных
func (p *PGXDB) UserRegistration(user domain.User) (id int, err error) {
	tx, err := p.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	_, err = tx.Exec("INSERT INTO gk_users (login, email, password, saltb64, created_at) VALUES ($1, $2, $3, $4, current_timestamp);", user.Login, user.Email, user.Password, user.SaltB64)
	row := tx.QueryRow("SELECT MAX(id) FROM gk_users;")
	err = row.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

// UserAuthorization получает пароль, соль и хэш крипто пароля по логину пользователя
func (p *PGXDB) UserAuthorization(login string) (int, string, string, string, error) {
	rows := p.DB.QueryRow("SELECT id, password, saltb64, COALESCE(secret_password_hash, '') FROM gk_users WHERE login = $1;", login)
	id := -1
	var password, saltb64, spasshash string
	err := rows.Scan(&id, &password, &saltb64, &spasshash)
	if err != nil {
		return -1, "", "", "", err
	}
	return id, password, saltb64, spasshash, nil
}

// SaveUserPasswordHash сохраняет хэш крипто пароля пользователя
func (p *PGXDB) SaveUserPasswordHash(session, hash string) error {
	_, err := p.DB.Exec("UPDATE gk_users SET secret_password_hash = $1 WHERE id = (SELECT users_id FROM gk_sessions WHERE session = $2);", hash, session)
	return err
}
