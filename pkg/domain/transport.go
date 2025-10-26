// Package domain описывает форматы сущностей всего проекта
package domain

import "net/http"

type (
	// Response описывает формат ответа сервера на клиете
	Response struct {
		JSONData   string
		Cookies    []*http.Cookie
		StatusCode int
		Err        error
	}
)
