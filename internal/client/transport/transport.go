// Package transport реализует сервис для объединения нескольких способов передачи данных до сервера
package transport

import (
	"gophkeeper/internal/client/transport/clienthttps"
	"gophkeeper/pkg/domain"

	log "github.com/sirupsen/logrus"
)

// структуры
type (
	// Client хранит все способы передачи данных
	Client struct {
		TransportServices []Transporter
		Sender            int
		Logger            *log.Logger
	}
	// Transporter интерфейс для подключения сервисов передачи данных реализующих его
	Transporter interface {
		Registration(*domain.User) (string, error)
		Authorization(*domain.User) (string, error)
		TOTPConfirmation(*domain.User, string) (string, error)
		TOTPUpdate(*domain.User, string) (string, error)
		SaveHashSecretPassword(domain.User, string) (string, error)
		SyncData(domain.SyncPayload) (domain.SyncPayload, string, error)
	}
)

// NewTransporter инициализирует новый сервис передачи данных
func NewTransporter(serverHTTPS, serverGRPC string, logger *log.Logger) *Client {
	cl := &Client{Logger: logger}
	cl.SetServers(serverHTTPS, serverGRPC)
	return cl
}

// SetServers устанавливает непосредственных сетевых клиентов
func (t *Client) SetServers(serverHTTPS, serverGRPC string) error {
	if serverHTTPS != "" {
		t.TransportServices = append(t.TransportServices, clienthttps.New(serverHTTPS, t.Logger))
		t.Sender = len(t.TransportServices) - 1
	}
	return nil
}

// Registration промежуточная функция регистрации
func (t *Client) Registration(user *domain.User) (textErr string, err error) {
	return t.TransportServices[t.Sender].Registration(user)
}

// Authorization промежуточная функция авторизации
func (t *Client) Authorization(user *domain.User) (textErr string, err error) {
	return t.TransportServices[t.Sender].Authorization(user)
}

// TOTPConfirmation промежуточная функция подтверждения одноразовым паролем
func (t *Client) TOTPConfirmation(user *domain.User, code string) (textErr string, err error) {
	return t.TransportServices[t.Sender].TOTPConfirmation(user, code)
}

// TOTPUpdate промежуточная функция обновления авторизации одноразового пароля
func (t *Client) TOTPUpdate(user *domain.User, code string) (textErr string, err error) {
	return t.TransportServices[t.Sender].TOTPUpdate(user, code)
}

// SaveHashSecretPassword промежуточная функция подтверждения одноразовым паролем
func (t *Client) SaveHashSecretPassword(user domain.User, hash string) (textErr string, err error) {
	return t.TransportServices[t.Sender].SaveHashSecretPassword(user, hash)
}

// SyncData промежуточная функция синхронизации данных
func (t *Client) SyncData(dataOut domain.SyncPayload) (domain.SyncPayload, string, error) {
	return t.TransportServices[t.Sender].SyncData(dataOut)
}
