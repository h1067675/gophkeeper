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

const (
	APIv1                  = "/api/v1"
	RouteUser              = APIv1 + "/user"
	RouteUserRegistration  = "/registration"
	RouteUserAuthorization = "/authorization"
	RouteUserTOTP          = "/totp"
	RouteUserTOTPUpdate    = "/totp-update"
	RouteUserSavePassHash  = "/save-hash"
	RouteSync              = APIv1 + "/sync"
)
