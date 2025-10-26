// Package mail отправляет письма пользователю по электронной почте
package mail

import (
	"fmt"
	"net/smtp"

	log "github.com/sirupsen/logrus"
)

// Mail описывает структуру почтового сервиса
type Mail struct {
	Auth     smtp.Auth
	SMTPHost string
	SMTPPort string
	From     string
	Password string
	Logger   *log.Logger
}

// New инициализирует почтовый сервис
func New(smtpHost, smtpPort, from, password string, logger *log.Logger) *Mail {
	auth := smtp.PlainAuth("", from, password, smtpHost)
	return &Mail{Auth: auth, SMTPHost: smtpHost, SMTPPort: smtpPort, From: from, Password: password, Logger: logger}
}

// SendEmail отправляет письмо
func (m *Mail) SendEmail(to []string, subject, body string) error {
	message := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
			"\r\n%s\r\n",
		m.From, to[0], subject, body))
	err := smtp.SendMail(m.SMTPHost+":"+m.SMTPPort, m.Auth, m.From, to, message)
	if err != nil {
		return err
	}
	return nil
}
