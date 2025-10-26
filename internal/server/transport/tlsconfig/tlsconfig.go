// Package tlsconfig реализует TLS конфигурацию сервера,
// при этом на локале использует самоподписанный сертификат,
// а в продакшене автоматический сертификат от Let's Encrypt через autocert.Manager
package tlsconfig

import (
	"crypto/tls"

	"golang.org/x/crypto/acme/autocert"
)

// LoadTLSConfig возвращает TLS-конфиг:
// - для localhost → самоподписанный сертификат (cert.pem + key.pem)
// - для реального домена → autocert.Manager (Let's Encrypt)
func LoadTLSConfig(domain string) (*tls.Config, error) {
	if domain == "localhost" || domain == "127.0.0.1" {
		// Загружаем самоподписанный сертификат
		cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
		if err != nil {
			return nil, err
		}
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
		}, nil
	}

	// Получаем сертификат от
	m := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}
	return m.TLSConfig(), nil
}
