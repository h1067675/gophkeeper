// Модуль servergrpc реализует сервер gRPC
//
// TODO: реализовать сервер
package servergrpc

import (
	"context"
	"crypto/tls"
	"gophkeeper/internal/server/gophkeeper"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// структуры
type (

	// Server отвечает за маршрутизацию
	Server struct {
		Server     *grpc.Server
		Listen     net.Listener
		GophKeeper *gophkeeper.Service
		Logger     *log.Logger
	}
)

// NewServer инициализирует gRPC сервер
func NewServer(runAddress string, service *gophkeeper.Service, logger *log.Logger) *Server {
	return &Server{Logger: logger}
}

// StartServer запускает gRPC сервер
func (s *Server) StartServer(ctx context.CancelFunc, tls *tls.Config) error {
	// fmt.Printf("GRPC server running at %v\n", s.Server)

	return nil
}

// StopServer останавливает gRPC сервер
func (s *Server) StopServer(ctx context.Context) error {

	// s.Server.Stop()
	// if err := s.Listen.Close(); err != nil {
	// 	return err
	// }
	return nil
}
