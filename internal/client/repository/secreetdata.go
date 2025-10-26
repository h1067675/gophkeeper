// Package repository реализует хранение локальных данных в SQLlite базе данных
package repository

import (
	"database/sql"
	"time"

	"gophkeeper/pkg/domain"
	"gophkeeper/pkg/errors"
)

// GetPasswordsList получает список паролей из базы данных как для показа пользователю так и
// для синхронизации с сервером если установлено время последней синхронизации
func (s *SQLLite) GetPasswordsList(from time.Time) (passwords []domain.Password, err error) {
	var rows *sql.Rows
	if from.Compare(time.Time{}) != 0 {
		rows, err = s.Query("SELECT id, server_id, login, password, domain, created_at, changed_at, description, deleted  FROM gk_passwords WHERE server_id < 0 OR changed_at > ?;", from)
	} else {
		rows, err = s.Query("SELECT id, server_id, login, password, domain, created_at, changed_at, description, deleted FROM gk_passwords WHERE deleted <> TRUE;")
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return passwords, errors.ErrSQLNoRows
		}
		return passwords, err
	}
	defer rows.Close()

	for rows.Next() {
		pass := domain.Password{}
		err := rows.Scan(&pass.LocalID,
			&pass.ID,
			&pass.Login,
			&pass.Password,
			&pass.Domain,
			&pass.CreatedAt,
			&pass.ChangedAt,
			&pass.Description,
			&pass.Deleted,
		)
		if err != nil {
			return nil, err
		}

		passwords = append(passwords, pass)
	}
	return
}

// GetCardsList получает список банковских карт из базы данных как для показа пользователю так и
// для синхронизации с сервером если установлено время последней синхронизации
func (s *SQLLite) GetCardsList(from time.Time) (cards []domain.Card, err error) {
	var rows *sql.Rows
	if from.Compare(time.Time{}) != 0 {
		rows, err = s.Query("SELECT id, server_id, bank, number, month, year, cvc, holder, created_at, changed_at, description, deleted FROM gk_bank_cards WHERE server_id < 0 OR changed_at > ?;", from)
	} else {
		rows, err = s.Query("SELECT id, server_id, bank, number, month, year, cvc, holder, created_at, changed_at, description, deleted FROM gk_bank_cards WHERE deleted <> TRUE;")
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cards, errors.ErrSQLNoRows
		}
		return cards, err
	}
	defer rows.Close()

	for rows.Next() {
		card := domain.Card{}
		err := rows.Scan(&card.LocalID,
			&card.ID,
			&card.Bank,
			&card.Number,
			&card.Month,
			&card.Year,
			&card.CVC,
			&card.Holder,
			&card.CreatedAt,
			&card.ChangedAt,
			&card.Description,
			&card.Deleted,
		)
		if err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}
	return
}

// GetTextDataList получает список текстовых данных из базы данных как для показа пользователю так и
// для синхронизации с сервером если установлено время последней синхронизации
func (s *SQLLite) GetTextDataList(from time.Time) (texts []domain.Text, err error) {
	t := from.Format(time.RFC3339)
	var rows *sql.Rows
	if from.Compare(time.Time{}) != 0 {
		rows, err = s.Query("SELECT id, server_id, name, data, created_at, changed_at, description, deleted FROM gk_text_data WHERE server_id < 0 OR changed_at > ?;", t)
	} else {
		rows, err = s.Query("SELECT id, server_id, name, data, created_at, changed_at, description, deleted FROM gk_text_data WHERE deleted <> TRUE;")
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return texts, errors.ErrSQLNoRows
		}
		return texts, err
	}
	defer rows.Close()

	for rows.Next() {
		txt := domain.Text{}
		err := rows.Scan(&txt.LocalID,
			&txt.ID,
			&txt.Name,
			&txt.Data,
			&txt.CreatedAt,
			&txt.ChangedAt,
			&txt.Description,
			&txt.Deleted,
		)
		if err != nil {
			return nil, err
		}

		texts = append(texts, txt)
	}
	return
}

// GetBinaryDataList получает список бинарных данных из базы данных как для показа пользователю так и
// для синхронизации с сервером если установлено время последней синхронизации
func (s *SQLLite) GetBinaryDataList(from time.Time) (binarys []domain.Binary, err error) {
	var rows *sql.Rows
	if from.Compare(time.Time{}) != 0 {
		rows, err = s.Query("SELECT id, server_id, name, data, created_at, changed_at, description, deleted FROM gk_binary_data WHERE server_id < 0 OR changed_at > ?;", from)
	} else {
		rows, err = s.Query("SELECT id, server_id, name, data, created_at, changed_at, description, deleted FROM gk_binary_data WHERE deleted <> TRUE;")
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return binarys, errors.ErrSQLNoRows
		}
		return binarys, err
	}
	defer rows.Close()

	for rows.Next() {
		bin := domain.Binary{}
		err := rows.Scan(&bin.LocalID,
			&bin.ID,
			&bin.Name,
			&bin.Data,
			&bin.CreatedAt,
			&bin.ChangedAt,
			&bin.Description,
			&bin.Deleted,
		)
		if err != nil {
			return nil, err
		}

		binarys = append(binarys, bin)
	}
	return
}

// GetPassword получает данные о пароле для отображения пользователю
func (s *SQLLite) GetPassword(id int) (pass domain.Password, err error) {
	row := s.QueryRow("SELECT id, server_id, login, password, domain, created_at, description FROM gk_passwords WHERE id = ? AND deleted <> TRUE;", id)
	err = row.Scan(&pass.LocalID,
		&pass.ID,
		&pass.Login,
		&pass.Password,
		&pass.Domain,
		&pass.CreatedAt,
		&pass.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pass, errors.ErrSQLNoRows
		}
		return pass, err
	}
	return
}

// GetCard получает данные о банковской карте для отображения пользователю
func (s *SQLLite) GetCard(id int) (card domain.Card, err error) {
	row := s.QueryRow("SELECT id, server_id, bank, number, month, year, cvc, holder, created_at, description FROM gk_bank_cards WHERE id = ? AND deleted <> TRUE;", id)
	err = row.Scan(&card.LocalID,
		&card.ID,
		&card.Bank,
		&card.Number,
		&card.Month,
		&card.Year,
		&card.CVC,
		&card.Holder,
		&card.CreatedAt,
		&card.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return card, errors.ErrSQLNoRows
		}
		return card, err
	}
	return
}

// GetTextData получает текстовые данные для отображения пользователю
func (s *SQLLite) GetTextData(id int) (txt domain.Text, err error) {
	row := s.QueryRow("SELECT id, server_id, name, data, created_at, description FROM gk_text_data WHERE id = ? AND deleted <> TRUE;", id)
	err = row.Scan(&txt.LocalID,
		&txt.ID,
		&txt.Name,
		&txt.Data,
		&txt.CreatedAt,
		&txt.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return txt, errors.ErrSQLNoRows
		}
		return txt, err
	}
	return
}

// GetBinaryData получает бинарные данные для отображения пользователю
func (s *SQLLite) GetBinaryData(id int) (bin domain.Binary, err error) {
	row := s.QueryRow("SELECT id, server_id, name, data, created_at, description FROM gk_binary_data WHERE id = ? AND deleted <> TRUE;", id)
	err = row.Scan(&bin.LocalID,
		&bin.ID,
		&bin.Name,
		&bin.Data,
		&bin.CreatedAt,
		&bin.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return bin, errors.ErrSQLNoRows
		}
		return bin, err
	}
	return
}

// NewPassword сохраняет новые данные пароля в базу данных
func (s *SQLLite) NewPassword(pass *domain.Password) (err error) {
	row, err := s.Exec(`INSERT INTO gk_passwords (server_id, login, password, domain, created_at, changed_at, description)
					VALUES (?, ?, ?, ?, ?, ?, ?);`,
		pass.ID, pass.Login, pass.Password, pass.Domain, pass.CreatedAt, pass.ChangedAt, pass.Description,
	)
	if err != nil {
		return err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return err
	}
	pass.LocalID = int(id)
	return nil

}

// NewCard сохраняет новую банковскую карту в базу данных
func (s *SQLLite) NewCard(card *domain.Card) (err error) {
	row, err := s.Exec(`INSERT INTO gk_bank_cards (server_id, bank, number, month, year, cvc, holder, created_at, changed_at, description)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		card.ID, card.Bank, card.Number, card.Month, card.Year, card.CVC, card.Holder, card.CreatedAt, card.ChangedAt, card.Description,
	)
	if err != nil {
		return err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return err
	}
	card.LocalID = int(id)
	return nil
}

// NewTextData сохраняет новые текстовые данные  в базу данных
func (s *SQLLite) NewTextData(txt *domain.Text) (err error) {
	row, err := s.Exec(`INSERT INTO gk_text_data (server_id, name, data, created_at, changed_at, description)
					VALUES (?, ?, ?, ?, ?, ?);`,
		txt.ID, txt.Name, txt.Data, txt.CreatedAt, txt.ChangedAt, txt.Description,
	)
	if err != nil {
		return err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return err
	}
	txt.LocalID = int(id)
	return nil

}

// NewBinaryData сохраняет новые бинарные данные  в базу данных
func (s *SQLLite) NewBinaryData(bin *domain.Binary) (err error) {
	row, err := s.Exec(`INSERT INTO gk_binary_data (server_id, name, data, created_at, changed_at, description)
					VALUES (?, ?, ?, ?, ?, ?);`,
		bin.ID, bin.Name, bin.Data, bin.CreatedAt, bin.ChangedAt, bin.Description,
	)
	if err != nil {
		return err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return err
	}
	bin.LocalID = int(id)
	return nil

}

// SavePassword сохраняет изммененные данные пароля в базу данных
func (s *SQLLite) SavePassword(pass domain.Password) (err error) {
	_, err = s.Exec(`UPDATE gk_passwords SET 
	server_id=?, login=?, password=?, domain=?, created_at=?, changed_at=?, description=?, deleted=?
	WHERE id =?;`,
		pass.ID, pass.Login, pass.Password, pass.Domain, pass.CreatedAt, pass.ChangedAt, pass.Description, pass.Deleted,
		pass.LocalID,
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveCard сохраняет изммененные данные банковской карты в базу данных
func (s *SQLLite) SaveCard(card domain.Card) (err error) {
	_, err = s.Exec(`UPDATE gk_bank_cards SET 
	server_id=?, bank=?, number=?, month=?, year=?, cvc=?, holder=?, created_at=?, changed_at=?, description=?, deleted=?
	WHERE id =?;`,
		card.ID, card.Bank, card.Number, card.Month, card.Year, card.CVC, card.Holder, card.CreatedAt, card.ChangedAt, card.Description, card.Deleted,
		card.LocalID,
	)
	if err != nil {
		return err
	}
	return
}

// SaveTextData сохраняет изммененные текстовые данные в базу данных
func (s *SQLLite) SaveTextData(txt domain.Text) (err error) {
	_, err = s.Exec(`UPDATE gk_text_data SET 
	server_id=?, name=?, data=?, created_at=?, changed_at=?, description=?, deleted=?
	WHERE id =?;`,
		txt.ID, txt.Name, txt.Data, txt.CreatedAt, txt.ChangedAt, txt.Description, txt.Deleted,
		txt.LocalID,
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveBinaryData сохраняет изммененные бинарные данные в базу данных
func (s *SQLLite) SaveBinaryData(bin domain.Binary) (err error) {
	_, err = s.Exec(`UPDATE gk_binary_data SET 
	server_id=?, name=?, data=?, created_at=?, changed_at=?, description=?, deleted=?
	WHERE id =?;`,
		bin.ID, bin.Name, bin.Data, bin.CreatedAt, bin.ChangedAt, bin.Description, bin.Deleted,
		bin.LocalID,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeletePassword помечает данные пароля в базе данных как удаленные
func (s *SQLLite) DeletePassword(pass domain.Password) (err error) {
	data := domain.Password{ID: pass.ID, LocalID: pass.LocalID, ChangedAt: pass.ChangedAt, Deleted: true}
	return s.SavePassword(data)
}

// DeleteCard помечает данные банковской карты в базе данных как удаленные
func (s *SQLLite) DeleteCard(card domain.Card) (err error) {
	data := domain.Card{ID: card.ID, LocalID: card.LocalID, ChangedAt: card.ChangedAt, Deleted: true}
	return s.SaveCard(data)
}

// DeleteTextData помечает текстовые данные в базе данных как удаленные
func (s *SQLLite) DeleteTextData(txt domain.Text) (err error) {
	data := domain.Text{ID: txt.ID, LocalID: txt.LocalID, ChangedAt: txt.ChangedAt, Deleted: true}
	return s.SaveTextData(data)
}

// DeleteBinaryData помечает бинарные данные в базе данных как удаленные
func (s *SQLLite) DeleteBinaryData(bin domain.Binary) (err error) {
	data := domain.Binary{ID: bin.ID, LocalID: bin.LocalID, ChangedAt: bin.ChangedAt, Deleted: true}
	return s.SaveBinaryData(data)
}

// SyncPassword сохраняет данные пароля полученные от сервера
func (s *SQLLite) SyncPassword(pass domain.Password) (err error) {
	if pass.LocalID > 0 {
		return s.SavePassword(pass)
	}
	return s.NewPassword(&pass)
}

// SyncCard сохраняет данные банковской карты полученные от сервера
func (s *SQLLite) SyncCard(card domain.Card) (err error) {
	if card.LocalID > 0 {
		return s.SaveCard(card)
	}
	return s.NewCard(&card)
}

// SyncText сохраняет текстовые данные полученные от сервера
func (s *SQLLite) SyncText(txt domain.Text) (err error) {
	if txt.LocalID > 0 {
		return s.SaveTextData(txt)
	}
	return s.NewTextData(&txt)
}

// SyncBinary сохраняет бинарные данные полученные от сервера
func (s *SQLLite) SyncBinary(bin domain.Binary) (err error) {
	if bin.LocalID > 0 {
		return s.SaveBinaryData(bin)
	}
	return s.NewBinaryData(&bin)
}
