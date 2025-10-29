// Package gophkeeper реализует основную бизнес логику сервера
// описывает интерфейсы для доступа к микросервисам
package gophkeeper

import (
	"context"
	"gophkeeper/pkg/domain"
	"gophkeeper/pkg/errors"
	"gophkeeper/pkg/random"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pquerna/otp"
)

// AuthorizationToken получает токен у модуля авторизации
func (g *Service) AuthorizationToken(token string) (string, error) {
	return g.Authorizer.GetSessionFromToken(token)
}

// CreateTokenSession получает сессию из токена от модуля авторизации
func (g *Service) CreateTokenSession(session string) (string, error) {
	return g.Authorizer.CreateTokenSession(session)
}

// UserRegistration осуществляет регистрацию пользователя
func (g *Service) UserRegistration(ctx context.Context, login string, password string, email string) (string, time.Time, string, string, error) {
	// валидируем данные регистрации
	user := domain.User{Login: login, Password: password, Email: email}
	if err := user.Validate(); err != nil {
		return "", time.Time{}, "", "", err
	}
	// хэшируем пароль пользователя
	hash, err := g.Crypter.HashArgon2id(user.Password)
	if err != nil {
		return "", time.Time{}, "", "", err
	}
	user.Password = hash
	// проверяем логин на уникальность
	if userid, err := g.Repository.UserGetByLogin(login); userid > -1 {
		return "", time.Time{}, "", "", errors.ErrUserAlreadyExist
	} else if err != nil {
		return "", time.Time{}, "", "", err
	}
	// проверяем email на уникальность
	if userid, err := g.Repository.UserGetByEmail(email); userid > -1 {
		return "", time.Time{}, "", "", errors.ErrUserEmailAlreadyExist
	} else if err != nil {
		return "", time.Time{}, "", "", err
	}
	// генерируем пользовательскую соль
	_, saltb64, err := g.Crypter.GenerateSalt()
	if err != nil {
		return "", time.Time{}, "", "", err
	}
	user.SaltB64 = saltb64
	// регистрируем пользователя
	user.ID, err = g.Repository.UserRegistration(user)
	if err != nil {
		return "", time.Time{}, "", "", err
	}
	// получаем уникальный код постоянной сессии, данная сессия не дает доступа к проверке или сохранению данных
	session, expire, err := g.GenerateSession(ctx, user.ID, false)
	if err != nil {
		return "", time.Time{}, "", "", err
	}
	// генерируем данные для регистрации TOTP Google Authorizator
	totp, err := g.GenerateTOTPSecret(ctx, session, user.Email)
	if err != nil {
		return "", time.Time{}, "", "", err
	}
	// возвращаем хэндлеру токен временной сессии срок действия и данные для TOTP
	token, err := g.CreateTokenSession(session)
	if err != nil {
		return "", time.Time{}, "", "", err
	}
	return token, expire, saltb64, totp.URL(), nil
}

// GenerateTOTPSecret получает данные о двухфакторной авторизации
func (g *Service) GenerateTOTPSecret(ctx context.Context, session string, email string) (*otp.Key, error) {
	// получаем уникальный код от Google Authorezation
	key, err := g.Authorizer.GenerateTOTPSecret(g.CompanyName, email)
	if err != nil {
		return nil, err
	}
	// сохраняем секрет для постоянной сессии пользователя в базу данных
	err = g.Repository.SessionSaveTOTPSecret(session, key.Secret())
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateSession генерирует сессию
func (g *Service) GenerateSession(ctx context.Context, userid int, allowed bool) (string, time.Time, error) {
	session, expire := g.Authorizer.GenerateSession(allowed)
	err := g.Repository.SessionSave(userid, session, expire, allowed)
	// устраняем возможность коллизии
	if err != nil {
		if errors.As(err, pgx.ErrTooManyRows) {
			return g.GenerateSession(ctx, userid, allowed)
		}
		return "", time.Time{}, err
	}

	return session, expire, nil
}

// ValidateTOTPCode проверяет код двухфакторной авторизации
func (g *Service) ValidateTOTPCode(ctx context.Context, session string, code string) (string, time.Time, error) {
	// запрашиваем секретный код TOTP в базе данных
	userID, secretKey, err := g.Repository.SessionGetTOTPSecret(session)
	if err != nil {
		return "", time.Time{}, err
	}
	// проверяем код пользователя
	err = g.Authorizer.ValidateTOTPCode(code, secretKey)
	if err != nil {
		return "", time.Time{}, err
	}
	// создаем сессию доступа сроком на 12 часов и отправляем данные клиенту
	session, expire, err := g.GenerateSession(ctx, userID, true)
	if err != nil {
		return "", time.Time{}, err
	}
	token, err := g.CreateTokenSession(session)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expire, nil
}

// UpdateTOTPRegistration обновляет данные авторизации в сервисе двухфакторной авторизации
func (g *Service) UpdateTOTPRegistration(ctx context.Context, session string, email string, code string) (string, error) {
	// запрашиваем секретный код TOTP в базе данных
	emailDB, codeDB, expireCode, err := g.Repository.SessionGetEmail(session)
	if err != nil {
		return "", err
	}
	if (code == "" && email == emailDB) || time.Now().After(expireCode) {
		code = random.RandCode()
		if err := g.Mailer.SendEmail([]string{email}, "Ваш код подтверждения для TOTP авторизации GophKeeper. Срок действия 12 часов.", code); err != nil {
			return "", err
		}
		if err := g.Repository.SessionSetEmailCode(session, code, time.Now().Add(12*time.Hour)); err != nil {
			return "", err
		}
		return "", nil
	} else if code != "" && codeDB == code && email == emailDB {

		// генерируем данные для регистрации TOTP Google Authorizator
		totp, err := g.GenerateTOTPSecret(ctx, session, email)
		if err != nil {
			return "", err
		}
		// возвращаем хэндлеру код временной сессии срок действия и данные для TOTP
		return totp.URL(), nil
	}
	return "", errors.ErrTOTPUpdate
}

// UserAuthorization осуществляет авторизацию пользователя по паролю
func (g *Service) UserAuthorization(ctx context.Context, login string, password string) (string, string, time.Time, string, string, error, error) {
	// валидируем данные регистрации
	user := domain.User{Login: login, Password: password}
	if err := user.Validate(); err != nil {
		if !errors.Is(err, errors.ErrUserInvalidEmail) {
			return "", "", time.Time{}, "", "", errors.ErrUserAuthorization, nil
		}
	}
	// сверяем данные пользователя с базой данных
	userID, hash, saltb64, spassHash, err := g.Repository.UserAuthorization(user.Login)
	if err != nil {
		return "", "", time.Time{}, "", "", errors.ErrUserAuthorization, err
	}
	// хэшируем пароль пользователя
	ok, err := g.Crypter.CompareHash(user.Password, hash)
	if err != nil || !ok {
		return "", "", time.Time{}, "", "", errors.ErrUserAuthorization, err
	}
	user.Password, user.ID = hash, userID
	// проверяем наличие кода постоянной сессии в базе данных
	session, secretKey, expire, err := g.Repository.SessionGet(user.ID)
	if err != nil {
		return "", "", time.Time{}, "", "", errors.ErrUserAuthorization, err
	}
	if session == "" || time.Now().After(expire) || secretKey == "" {
		return domain.APIv1 + domain.RouteUserTOTPUpdate, session, expire, "", "", errors.ErrTOTPExpired, err
	}
	token, err := g.CreateTokenSession(session)
	if err != nil {
		return "", "", time.Time{}, "", "", nil, err
	}
	if spassHash == "" {
		return domain.APIv1 + domain.RouteUserSavePassHash, token, expire, "", "", errors.ErrUserSecretPassword, nil
	}

	// возвращаем хэндлеру код временной сессии срок действия
	return "", token, expire, saltb64, spassHash, nil, nil
}

// UserSavePasswordHash сохраняет строку сессии защифрованную при помощи пароля шифрования пользователя и уникальной соли
func (g *Service) UserSavePasswordHash(ctx context.Context, allowedToken, hash string) error {
	err := g.Repository.SaveUserPasswordHash(allowedToken, hash)
	if err != nil {
		return err
	}

	return nil
}

// SyncData получает данные от пользователя и сохраняет изменения в базе данных,
// при этом указывает присвоенные ID и добавляет в структуру данные которых у пользователя нет
func (g *Service) SyncData(ctx context.Context, allowedToken string, dataIn domain.SyncPayload) (dataOut domain.SyncPayload, err error) {
	userID, expired, err := g.Repository.UserGetByToken(allowedToken)
	if err != nil {
		return dataOut, err
	}
	if time.Now().After(expired) {
		return dataIn, errors.ErrTOTPExpired
	}
	now := time.Now()
	for i := range dataIn.Passwords {
		dataIn.Passwords[i].UserID, dataIn.Passwords[i].ChangedAt = userID, now
		id, err := g.Repository.Upsert(dataIn.Passwords[i], dataIn.Passwords[i].TableName(), "id")
		if err != nil {
			return dataOut, err
		}
		dataIn.Passwords[i].ID = id
	}
	err = g.Repository.Select(&dataOut.Passwords, "SELECT * FROM "+domain.Password{}.TableName()+" WHERE users_id = $1 AND changed_at >= $2 AND changed_at < $3;", userID, dataIn.LastSync, now)
	if err != nil {
		return dataOut, err
	}
	dataOut.Passwords = append(dataOut.Passwords, dataIn.Passwords...)

	for i := range dataIn.Cards {
		dataIn.Cards[i].UserID, dataIn.Cards[i].ChangedAt = userID, now
		id, err := g.Repository.Upsert(dataIn.Cards[i], dataIn.Cards[i].TableName(), "id")
		if err != nil {
			return dataOut, err
		}
		dataIn.Cards[i].ID = id
	}
	err = g.Repository.Select(&dataOut.Cards, "SELECT * FROM "+domain.Card{}.TableName()+" WHERE users_id = $1 AND changed_at > $2 AND changed_at < $3;", userID, dataIn.LastSync, now)
	if err != nil {
		return dataOut, err
	}
	dataOut.Cards = append(dataOut.Cards, dataIn.Cards...)

	for i := range dataIn.Texts {
		dataIn.Texts[i].UserID, dataIn.Texts[i].ChangedAt = userID, now
		id, err := g.Repository.Upsert(dataIn.Texts[i], dataIn.Texts[i].TableName(), "id")
		if err != nil {
			return dataOut, err
		}
		dataIn.Texts[i].ID = id
	}
	err = g.Repository.Select(&dataOut.Texts, "SELECT * FROM "+domain.Text{}.TableName()+" WHERE users_id = $1 AND changed_at > $2 AND changed_at < $3;", userID, dataIn.LastSync, now)
	if err != nil {
		return dataOut, err
	}
	dataOut.Texts = append(dataOut.Texts, dataIn.Texts...)

	for i := range dataIn.Binaries {
		dataIn.Binaries[i].UserID, dataIn.Binaries[i].ChangedAt = userID, now
		id, err := g.Repository.Upsert(dataIn.Binaries[i], dataIn.Binaries[i].TableName(), "id")
		if err != nil {
			return dataOut, err
		}
		dataIn.Binaries[i].ID = id
	}
	err = g.Repository.Select(&dataOut.Binaries, "SELECT * FROM "+domain.Binary{}.TableName()+" WHERE users_id = $1 AND changed_at > $2 AND changed_at < $3;", userID, dataIn.LastSync, now)
	if err != nil {
		return dataOut, err
	}
	dataOut.Binaries = append(dataOut.Binaries, dataIn.Binaries...)
	dataOut.LastSync = now

	return dataOut, nil
}
