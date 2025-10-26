// Package serverhttps реализует сервер HTTPS
package serverhttps

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"gophkeeper/internal/server/gophkeeper"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// структуры
type (
	// Server описывает сервер HTTPS
	Server struct {
		Server     *http.Server
		GophKeeper *gophkeeper.Service
		Logger     *log.Logger
		chi.Router
	}
)

// NewServer инициализирует HTTPS сервер
func NewServer(runAddress string, service *gophkeeper.Service, logger *log.Logger) *Server {
	var s = &Server{Logger: logger, GophKeeper: service}
	// Определяем HTTP сервер и указываем адрес и ручку
	s.Server = &http.Server{
		Addr: runAddress,
	}

	return s
}

// StartServer запускает HTTPS сервер
func (s *Server) StartServer(cancel context.CancelFunc, tls *tls.Config) error {
	var err error
	s.Server.TLSConfig = tls
	if err != nil {
		return err
	}
	s.Server.Handler = s.createRouting()
	// Запускаем сервер с HTTPS
	go func() {
		if err = s.Server.ListenAndServeTLS("", ""); err != nil {
			s.Logger.Debugf("HTTPS server stopped whith error %e", err)
			fmt.Printf("HTTPS server stopped whith error: %v\n", err)
			cancel()
		}
	}()
	fmt.Printf("HTTPS server running at %v\n", s.Server.Addr)
	return err
}

// StopServer останавливает HTTPS сервер
func (s *Server) StopServer(ctx context.Context) error {
	if err := s.Server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
