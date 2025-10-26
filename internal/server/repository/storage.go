// Package repository реализует хранение данных сервера в PostgeSQL базе данных
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Select получает несколько записей из базы данных для структуры с тегами db
func (p *PGXDB) Select(dest any, query string, args ...any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return errors.New("Select() dest must be pointer to slice")
	}

	slice := v.Elem()
	elemType := slice.Type().Elem()

	rows, err := p.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		elem := reflect.New(elemType).Interface()
		if err := scanStruct(rows, elem); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, reflect.ValueOf(elem).Elem()))
	}

	return nil
}

// Upsert сохранят данные в базе в существующую запись если ID > 0 или в новую если меньше структуры с тегами db
func (p *PGXDB) Upsert(obj any, table string, conflictCol string) (int, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var cols []string
	var placeholders []string
	args := make([]any, 0, t.NumField())

	j := 1
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		col := f.Tag.Get("db")
		if col == "" || col == "-" {
			continue
		}
		if col == conflictCol && v.Field(i).Interface() == -1 {
			continue
		}
		cols = append(cols, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", j))
		j++
		args = append(args, v.Field(i).Interface())
	}

	// собираем список для DO UPDATE SET
	var updates []string
	for _, col := range cols {
		if col == conflictCol {
			continue
		}
		updates = append(updates, fmt.Sprintf("%s=excluded.%s", col, col))
	}

	// добавляем условие обновления только если новая запись свежее
	query := fmt.Sprintf(`
INSERT INTO %s (%s)
VALUES (%s)
ON CONFLICT(%s) DO UPDATE
SET %s
WHERE excluded.changed_at > %s.changed_at RETURNING id;
`, table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		conflictCol,
		strings.Join(updates, ", "),
		table,
	)

	var id int
	err := p.QueryRow(query, args...).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, err
}

// scanStruct связывает SQL-результат с полями структуры
func scanStruct(rows *sql.Rows, dest any) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	ptrs := make([]any, len(columns))
	fieldMap := map[string]int{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Tag.Get("db")
		if name == "" {
			name = field.Name
		}
		fieldMap[name] = i
	}

	for i, col := range columns {
		if idx, ok := fieldMap[col]; ok {
			ptrs[i] = v.Field(idx).Addr().Interface()
		} else {
			var tmp any
			ptrs[i] = &tmp
		}
	}

	return rows.Scan(ptrs...)
}
