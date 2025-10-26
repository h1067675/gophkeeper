// Package app обеспечивает запуск остальных сервисов и реализует функцию безопасного завершения работы
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophkeeper/internal/server/authorization"
	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/gophkeeper"
	"gophkeeper/internal/server/mail"
	"gophkeeper/internal/server/repository"
	"gophkeeper/internal/server/transport"
	cryptoargon "gophkeeper/pkg/crypto"
	"gophkeeper/pkg/logger"

	log "github.com/sirupsen/logrus"
)

// structures
type (
	// Application описывает основной поток сервера
	Application struct {
		Config *config.Config
		Server *transport.Server
		Logger *log.Logger
	}
)

var (
	// Определяем глобальные переменные для вывода версии сборки указаннной при компиляции
	buildVersion string
	buildDate    string
	buildCommit  string
)

// NewApplication инициализирует сервет
func NewApplication(domain, httpsport, grpcport, databaseDNS, company, smtphost, smtpport, email string, dropDB bool) (*Application, error) {
	// Инициализируем логгер.
	logger := logger.InitializeLogger(&log.JSONFormatter{}, log.InfoLevel, os.Stdout)
	// Получаем данные о версии и печатаем в os.Stout
	versionControl()
	// Инициализируем конфигурацию
	conf, err := config.InitializeConfigurer(domain, httpsport, grpcport, databaseDNS, company, smtphost, smtpport, email, logger)
	if err != nil {
		logger.Debug("Сonfiguration initialization error.\nGophKeeper Server is stopped.\n")
		return nil, err
	}
	// Запускаем хранилище
	repo, err := repository.New(conf.DatabaseDNS.String(), logger, dropDB)
	if err != nil {
		logger.Debugf("Errors occurred while connecting to the database.\nDatabase DNS - %s\nGophKeeper Server is stopped.\n",
			conf.DatabaseDNS.String())
		return nil, err
	}
	// Загружаем модуль авторизации
	auth, err := authorization.Init(conf.SecretKey.String(), logger)
	if err != nil {
		logger.Debug("Authorization module initialization error\nGophKeeper Server is stopped.\n")
		return nil, err
	}
	// Загружаем модуль шифрования
	crypt := cryptoargon.New(logger)
	// Загружаем почтовый модуль
	mail := mail.New(conf.SMTPHost.String(), conf.SMTPPort.String(), conf.Email.String(), conf.EmailPassword.String(), logger)
	// Создаем GophKeeperService
	gophKeeper, err := gophkeeper.NewService(conf.CompanyName.String(), repo, auth, crypt, mail, logger)
	if err != nil {
		logger.Debugf("Error starting service GophKeeper.\nGophKeeper Server is stopped.\n")
		return nil, err
	}
	// Создаем сервер
	serv := transport.NewServer(conf.Domain.String(), conf.HTTPSPort.String(), conf.GRPCPort.String(), gophKeeper, logger)
	// Создаем Application и делаем связи
	app := &Application{Server: serv, Config: conf, Logger: logger}

	return app, nil
}

// RunApplication запускает сервер и обеспечивает безопасное завершение работы после получения
// сигналов os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT
func (a *Application) RunApplication() (chan struct{}, context.CancelFunc, error) {
	// Создаем канал для ожидания сигнала от системы на окончание работы
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	idleConnsClosed := a.waitSysSignals(ctx)
	// Запускаем сервер
	err := a.Server.Start(cancel)
	if err != nil {
		a.Logger.WithError(err).Infof("Server startup error\nGophKeeper Server is stopped.\n")
		return nil, cancel, err
	}

	return idleConnsClosed, cancel, nil
}

// versionControl получает глобальные переменные версии и выводит в консоль
func versionControl() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("GophKeeper Server\n")
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

// WaitSysSignals определяет логику для отслеживания сигналов завершения работы приложения и
// реализует процесс мягкого завершения работы приложения.
func (a *Application) waitSysSignals(ctx context.Context) chan struct{} {
	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		<-ctx.Done()
		// создаем контекст с таймаутом на завершение операций сервером
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		// получили сигнал, запускаем процедуру graceful shutdown сервера
		if err := a.Server.ShutDown(ctx); err != nil {
			a.Logger.Debug(err)
		} else {
			a.Logger.Debug("Server is stopped.")
			fmt.Printf("Server is stopped.\n")
		}
		// сообщаем основному потоку и логгируем, что все сетевые соединения обработаны и закрыты
		a.Logger.Debug("GophKeeper has shutdown in graceful mode.")
		close(idleConnsClosed)
	}()
	return idleConnsClosed
}
