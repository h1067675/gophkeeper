// Package serverhttps реализует сервер HTTPS
package serverhttps

import (
	"gophkeeper/pkg/domain"

	"github.com/go-chi/chi"
)

// константы отписывающие данные передаваемые в context следующим хэндлерам
const (
	sessionID key = iota
	sessionAllowed
)

// key необходим для передачи данных через context
type key int

// createRouting делает маршрутизацию к хандлерам.
func (s *Server) createRouting() chi.Router {
	// Создаем chi роутер
	s.Router = chi.NewRouter()

	// Добавляем все функции middleware
	s.Router.Use(s.AuthorizationTokenMiddleware)
	s.Router.Use(s.LoggingMiddleware)
	s.Router.Use(s.CompressMiddleware)

	// Делаем маршрутизацию
	s.Router.Route("/", func(c chi.Router) {
		c.Route(domain.RouteUser, func(c chi.Router) {
			c.Post(domain.RouteUserRegistration, s.UserRegistrationHandler)     // POST запрос на регистрацию
			c.Post(domain.RouteUserAuthorization, s.UserAuthorizationHandler)   // POST запрос на авторизацию
			c.Post(domain.RouteUserTOTP, s.TOTPCheckHandler)                    // POST запрос на TOTP подтверждение регистрации
			c.Post(domain.RouteUserTOTPUpdate, s.TOTPUpdateHandler)             // POST запрос на обновление TOTP регистрации
			c.Post(domain.RouteUserSavePassHash, s.UserSavePasswordHashHandler) // POST запрос на обновление TOTP регистрации
		})
		c.Post(domain.RouteSync, s.SyncDataHandler) // POST запрос на синхронизацию данных
	})

	return s.Router
}
