// Package gophkeeper реализует основную бизнес логику приложения
// описывает интерфейсы для доступа к микросервисам
package gophkeeper

import (
	"gophkeeper/pkg/domain"
	"gophkeeper/pkg/errors"
	"time"
)

// SaveServerConficuration сохраняет конфигурацию предоставленную пользователем
func (g *Service) SaveServerConficuration(serverAddress, grpcport, httpport string) (err error) {
	err = g.Configurer.SetConfig(serverAddress, grpcport, httpport)
	if err != nil {
		return err
	}
	err = g.Transporter.SetServers(g.Configurer.GetGRPCAddress(), g.Configurer.GetHTTPSAddress())
	if err != nil {
		return err
	}
	return nil
}

// GetSession получает данные о сессии
func (g *Service) GetSession() string {
	return g.Configurer.GetUserSession()
}

// GetSaltB64 получает шифровальную соль
func (g *Service) GetSaltB64() string {
	return g.Configurer.GetUserSaltB64()
}

// GetSPassHas получает хэш криптопароля
func (g *Service) GetSPassHas() string {
	return g.Configurer.GetUserSPassHash()
}

// Registration осуществляет регистрацию пользователя
func (g *Service) Registration() (domain.Screen, string) {
	if err := g.User.Validate(); err != nil {
		return domain.ScreenRegister, err.Error()
	}
	textErr, err := g.Transporter.Registration(&g.User)
	if err != nil {
		return domain.ScreenRegister, err.Error()
	}
	if err := g.Configurer.SaveUserSession(g.User.Session); err != nil {
		g.Logger.Debug(err)
	}
	if err := g.Configurer.SaveUserSaltB64(g.User.SaltB64); err != nil {
		g.Logger.Debug(err)
	}
	return domain.ScreenTOTP, textErr
}

// Authorization осуществляет авторизацию пользователя
func (g *Service) Authorization() (domain.Screen, string) {
	if err := g.User.Validate(); err != nil && !errors.Is(err, errors.ErrUserInvalidEmail) {
		return domain.ScreenLogin, " Ошибка! Неправильный логин или пароль. "
	}
	next := domain.ScreenTOTP
	textErr, err := g.Transporter.Authorization(&g.User)
	if errors.Is(err, errors.ErrTOTPExpired) {
		textErr += " Истек срок авторизации Goggle Authorizate. "
	} else if errors.Is(err, errors.ErrUserSecretPassword) {
		next = domain.ScreenSecretPassword
		textErr += " Необходимо установить крипто-пароль."
	} else if err != nil || (g.User.Session == "" && textErr != "") {
		return domain.ScreenLogin, textErr
	}
	if err := g.Configurer.SaveUserSession(g.User.Session); err != nil {
		g.Logger.Debug(err)
	}
	if err := g.Configurer.SaveUserSaltB64(g.User.SaltB64); err != nil {
		g.Logger.Debug(err)
	}
	if err := g.Configurer.SaveUserSPassHash(g.User.SPassHash); err != nil {
		g.Logger.Debug(err)
	}
	return next, textErr
}

// TOTPConfirmation осуществляет подтверждение вход пользователя одноразовым паролем
func (g *Service) TOTPConfirmation(code string) (domain.Screen, string) {
	textErr, err := g.Transporter.TOTPConfirmation(&g.User, code)
	if err != nil {
		return domain.ScreenTOTP, err.Error()
	}
	if g.User.Token == "" {
		return domain.ScreenTOTP, textErr
	}
	return domain.ScreenSecretPassword, textErr
}

// TOTPUpdate создает новую авторизацию или восстанавливает потерянную авторизацию в Google Authorizator
func (g *Service) TOTPUpdate(code string) (domain.Screen, string) {
	textErr, err := g.Transporter.TOTPUpdate(&g.User, code)
	if err != nil {
		return domain.ScreenTOTP, err.Error()
	}
	return domain.ScreenTOTP, textErr
}

// SaveHashSecretPassword отправляет на сервер хэш шифровального пароля пользователя с
// целью передачи его на другие устройства и дальнейшей синхронизации.
// В данный хэш шифруется сессия пользователя на основании пароля
// шифрования и соли сгенерированной специльно для пользователя
func (g *Service) SaveHashSecretPassword() (domain.Screen, string) {
	nonceAndCipherData, err := g.Crypter.Encrypt([]byte(g.User.Session), g.User.ClientPass, g.User.SaltB64)
	if err != nil {
		return domain.ScreenSecretPassword, "Ошибка создания пароля"
	}
	textErr, err := g.Transporter.SaveHashSecretPassword(g.User, nonceAndCipherData)
	if err != nil {
		return domain.ScreenSecretPassword, err.Error()
	}
	if textErr != "" {
		return domain.ScreenSecretPassword, textErr
	}
	g.Configurer.SaveUserSPassHash(nonceAndCipherData)
	return domain.ScreenHome, ""
}

// CheckSecretPassword расшифровывает хэш пароля и сравнивает результат с сессией пользователя
func (g *Service) CheckSecretPassword() (domain.Screen, string) {
	if g.User.SPassHash == "" {
		g.User.SPassHash = g.GetSPassHas()
	}
	sesssion, err := g.Crypter.Decrypt(g.User.SPassHash, g.User.ClientPass, g.User.SaltB64)
	if err != nil {
		return domain.ScreenSecretPassword, err.Error()
	}
	if string(sesssion) != g.User.Session {
		return domain.ScreenSecretPassword, "Не верный пароль"
	}

	return domain.ScreenHome, ""
}

// QRMediumToString получает QR код из строки авторизации Google Auth для отображения пользователю
func (g *Service) QRMediumToString(qrString string) string {
	qr, err := g.QRGenerator.QRMediumToString(qrString)
	if err != nil {
		return "ошибка создания QR-кода"
	}
	return qr
}

// SessionExit удаляет данные пользователя с текущей рабочей станции
func (g *Service) SessionExit() string {
	g.User = domain.User{}
	if err := errors.Join(g.Configurer.DeleteUserSession(),
		g.Configurer.DeleteUserSaltB64(),
		g.Configurer.DeleteUserSPassHash(),
		g.Repository.DropMigrations()); err != nil {
		return "Не удалось произвести выход"
	}
	return "Сессия успешно сброшена"
}

// GetPasswordsList получает список паролей для отображению пользователю
func (g *Service) GetPasswordsList() ([]domain.Password, string) {
	passwords, err := g.Repository.GetPasswordsList(time.Time{})
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return []domain.Password{}, "Данные отсутствуют"
		}
		return []domain.Password{}, err.Error()
	}
	return passwords, ""
}

// GetCardsList получает список банковских карт для отображению пользователю
func (g *Service) GetCardsList() ([]domain.Card, string) {
	cards, err := g.Repository.GetCardsList(time.Time{})
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return []domain.Card{}, "Данные отсутствуют"
		}
		return []domain.Card{}, err.Error()
	}
	return cards, ""
}

// GetTextDataList получает список текстовых данных для отображению пользователю
func (g *Service) GetTextDataList() ([]domain.Text, string) {
	texts, err := g.Repository.GetTextDataList(time.Time{})
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return []domain.Text{}, "Данные отсутствуют"
		}
		return []domain.Text{}, err.Error()
	}
	return texts, ""
}

// GetBinaryDataList получает список бинарных данных для отображению пользователю
func (g *Service) GetBinaryDataList() ([]domain.Binary, string) {
	binarys, err := g.Repository.GetBinaryDataList(time.Time{})
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return []domain.Binary{}, "Данные отсутствуют"
		}
		return []domain.Binary{}, err.Error()
	}
	return binarys, ""
}

// DeletePassword помечает пароль в локальной базе как удаленный и стирает из записи конфиденциальные данные
func (g *Service) DeletePassword(pass domain.Password) string {
	pass.ChangedAt = time.Now()
	err := g.Repository.DeletePassword(pass)
	if err != nil {
		return err.Error()
	}
	return ""
}

// DeleteCard помечает банковскую карту в локальной базе как удаленную и стирает из записи конфиденциальные данные
func (g *Service) DeleteCard(card domain.Card) string {
	card.ChangedAt = time.Now()
	err := g.Repository.DeleteCard(card)
	if err != nil {
		return err.Error()
	}
	return ""
}

// DeleteTextData помечает текстовые данные в локальной базе как удаленные и стирает из записи конфиденциальные данные
func (g *Service) DeleteTextData(text domain.Text) string {
	text.ChangedAt = time.Now()
	err := g.Repository.DeleteTextData(text)
	if err != nil {
		return err.Error()
	}
	return ""
}

// DeleteBinaryData помечает бинарные данные в локальной базе как удаленные и стирает из записи конфиденциальные данные
func (g *Service) DeleteBinaryData(bin domain.Binary) string {
	bin.ChangedAt = time.Now()
	err := g.Repository.DeleteBinaryData(bin)
	if err != nil {
		return err.Error()
	}
	return ""
}

// GetPassword получает пароль из локальной базы данных и расшифровывает секретную информацию
func (g *Service) GetPassword(id int) (domain.Password, string) {
	pass, err := g.Repository.GetPassword(id)
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return domain.Password{}, "Данные отсутствуют"
		}
		return domain.Password{}, err.Error()
	}
	if pass.Password != "" {
		data, err := g.Crypter.Decrypt(pass.Password, g.User.ClientPass, g.User.SaltB64)
		if err != nil {
			return domain.Password{}, "Ошибка получения данных"
		}
		pass.Password = string(data)
	}

	return pass, ""
}

// GetCard получает банковскую карту из локальной базы данных и расшифровывает секретную информацию
func (g *Service) GetCard(id int) (domain.Card, string) {
	card, err := g.Repository.GetCard(id)
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return domain.Card{}, "Данные отсутствуют"
		}
		return domain.Card{}, err.Error()
	}
	if card.CVC != "" {
		data, err := g.Crypter.Decrypt(card.CVC, g.User.ClientPass, g.User.SaltB64)
		if err != nil {
			return domain.Card{}, "Ошибка получения данных"
		}
		card.CVC = string(data)
	}
	return card, ""
}

// GetTextData получает текстовые данные из локальной базы данных и расшифровывает секретную информацию
func (g *Service) GetTextData(id int) (domain.Text, string) {
	txt, err := g.Repository.GetTextData(id)
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return domain.Text{}, "Данные отсутствуют"
		}
		return domain.Text{}, err.Error()
	}
	if txt.Data != "" {
		data, err := g.Crypter.Decrypt(txt.Data, g.User.ClientPass, g.User.SaltB64)
		if err != nil {
			return domain.Text{}, "Ошибка получения данных"
		}
		txt.Data = string(data)
	}
	return txt, ""
}

// GetBinaryData получает текстовые данные из локальной базы данных и расшифровывает секретную информацию
func (g *Service) GetBinaryData(id int) (domain.Binary, string) {
	bin, err := g.Repository.GetBinaryData(id)
	if err != nil {
		if errors.Is(err, errors.ErrSQLNoRows) {
			return domain.Binary{}, "Данные отсутствуют"
		}
		return domain.Binary{}, err.Error()
	}
	if bin.Data != "" {
		data, err := g.Crypter.Decrypt(bin.Data, g.User.ClientPass, g.User.SaltB64)
		if err != nil {
			return domain.Binary{}, "Ошибка получения данных"
		}
		bin.Data = string(data)
	}
	return bin, ""
}

// SavePassword сохраняет пароль в локальную базу данных и шифрует секретную информацию
func (g *Service) SavePassword(pass *domain.Password) string {
	nonceAndCipherData, err := g.Crypter.Encrypt([]byte(pass.Password), g.User.ClientPass, g.User.SaltB64)
	if err != nil {
		return err.Error()
	}
	pass.Password = nonceAndCipherData
	pass.ChangedAt = time.Now()
	if pass.LocalID == -1 {
		if err := g.Repository.NewPassword(pass); err != nil {
			return err.Error()
		}
	} else {
		if err := g.Repository.SavePassword(*pass); err != nil {
			return err.Error()
		}
	}

	return ""
}

// SaveCard сохраняет банковскую карту в локальную базу данных и шифрует секретную информацию
func (g *Service) SaveCard(card *domain.Card) string {
	nonceAndCipherData, err := g.Crypter.Encrypt([]byte(card.CVC), g.User.ClientPass, g.User.SaltB64)
	if err != nil {
		return err.Error()
	}
	card.CVC = nonceAndCipherData
	card.ChangedAt = time.Now()
	if card.LocalID == -1 {
		if err := g.Repository.NewCard(card); err != nil {
			return err.Error()
		}
	} else {
		if err := g.Repository.SaveCard(*card); err != nil {
			return err.Error()
		}
	}

	return ""
}

// SaveTextData сохраняет текстовые данные в локальную базу данных и шифрует секретную информацию
func (g *Service) SaveTextData(txt *domain.Text) string {
	nonceAndCipherData, err := g.Crypter.Encrypt([]byte(txt.Data), g.User.ClientPass, g.User.SaltB64)
	if err != nil {
		return err.Error()
	}
	txt.Data = nonceAndCipherData
	txt.ChangedAt = time.Now()
	if txt.LocalID == -1 {
		if err := g.Repository.NewTextData(txt); err != nil {
			return err.Error()
		}
	} else {
		if err := g.Repository.SaveTextData(*txt); err != nil {
			return err.Error()
		}
	}
	return ""
}

// SaveBinaryData сохраняет бинарные данные данные в локальную базу данных и шифрует секретную информацию
func (g *Service) SaveBinaryData(bin *domain.Binary) string {
	nonceAndCipherData, err := g.Crypter.Encrypt([]byte(bin.Data), g.User.ClientPass, g.User.SaltB64)
	if err != nil {
		return err.Error()
	}
	bin.Data = nonceAndCipherData
	bin.ChangedAt = time.Now()
	if bin.LocalID == -1 {
		if err := g.Repository.NewBinaryData(bin); err != nil {
			return err.Error()
		}
	} else {
		if err := g.Repository.SaveBinaryData(*bin); err != nil {
			return err.Error()
		}
	}

	return ""
}

// SyncData синхронизирует локальную базу данных с сервером.
// Синхронизация основывается на дате последней синхронизации все что было изменено
// после нее передается на сервер и с сервера получаются все данные которые поступили после даты синхронизации
func (g *Service) SyncData() (string, error) {
	t, err := g.Repository.GetLastSync()
	if err != nil {
		return err.Error(), err
	}
	pass, err := g.Repository.GetPasswordsList(t)
	if err != nil {
		return err.Error(), err
	}
	cards, err := g.Repository.GetCardsList(t)
	if err != nil {
		return err.Error(), err
	}
	texts, err := g.Repository.GetTextDataList(t)
	if err != nil {
		return err.Error(), err
	}
	bins, err := g.Repository.GetBinaryDataList(t)
	if err != nil {
		return err.Error(), err
	}
	syncOut := domain.SyncPayload{
		Passwords:    pass,
		Cards:        cards,
		Texts:        texts,
		Binaries:     bins,
		AllowedToken: g.User.Token,
		LastSync:     t,
	}
	syncIn, textErr, err := g.Transporter.SyncData(syncOut)
	if err != nil || textErr != "" {
		return textErr, err
	}
	if err := g.updateData(syncIn); err != nil {
		return err.Error(), err
	}

	return "Данные успешно синъронизированны с сервером", nil
}

// updateData реализует процесс сохранения полученных данных с сервера
func (g *Service) updateData(syncIn domain.SyncPayload) error {
	for _, e := range syncIn.Passwords {
		if err := g.Repository.SyncPassword(e); err != nil {
			return err
		}
	}
	for _, e := range syncIn.Cards {
		if err := g.Repository.SyncCard(e); err != nil {
			return err
		}
	}
	for _, e := range syncIn.Texts {
		if err := g.Repository.SyncText(e); err != nil {
			return err
		}
	}
	for _, e := range syncIn.Binaries {
		if err := g.Repository.SyncBinary(e); err != nil {
			return err
		}
	}
	if err := g.Repository.UpdateLastSync(syncIn.LastSync); err != nil {
		return err
	}
	return nil
}
