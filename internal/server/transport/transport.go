// Package transport объединяет слои серверов
package transport

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/server/gophkeeper"
	"gophkeeper/internal/server/transport/servergrpc"
	"gophkeeper/internal/server/transport/serverhttps"
	"gophkeeper/internal/server/transport/tlsconfig"

	log "github.com/sirupsen/logrus"
)

// структуры
type (
	// Server описывает структуру транспортного модуля
	Server struct {
		ServerHTTPS *serverhttps.Server
		ServerGRPC  *servergrpc.Server
		Logger      *log.Logger
		Domain      string
	}
)

// NewServer инициализирует транспортноый модуль
func NewServer(domain string, httpport string, grpcport string, service *gophkeeper.Service, logger *log.Logger) *Server {
	server := &Server{Logger: logger, Domain: domain}
	server.ServerHTTPS = serverhttps.NewServer(fmt.Sprintf("%s:%s", domain, httpport), service, logger)
	server.ServerGRPC = servergrpc.NewServer(fmt.Sprintf("%s:%s", domain, httpport), service, logger)

	return server
}

// Start зщапускает все серверы
func (s *Server) Start(cancel context.CancelFunc) error {
	tlsconf, err := tlsconfig.LoadTLSConfig(s.Domain)
	if err != nil {
		return err
	}
	if err := s.ServerHTTPS.StartServer(cancel, tlsconf); err != nil {
		return err
	}
	if err := s.ServerGRPC.StartServer(cancel, tlsconf); err != nil {
		return err
	}
	return nil
}

// ShutDown останавливает все серверы
func (s *Server) ShutDown(ctx context.Context) error {
	return errors.Join(s.ServerHTTPS.StopServer(ctx), s.ServerGRPC.StopServer(ctx))
}
