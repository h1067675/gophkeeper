// Package app обеспечивает запуск остальных сервисов и реализует функцию безопасного завершения работы
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/gophkeeper"
	"gophkeeper/internal/client/repository"
	"gophkeeper/internal/client/transport"
	"gophkeeper/internal/client/ui"
	cryptoargon "gophkeeper/pkg/crypto"
	"gophkeeper/pkg/logger"
	"gophkeeper/pkg/qrgenerator"

	log "github.com/sirupsen/logrus"
)

// structures
type (
	// Application реализует основной процесс
	Application struct {
		TUI    *ui.BubbleTeaUI
		Config *config.Config
		Logger *log.Logger
	}
)

var (
	// Определяем глобальные переменные для вывода версии сборки указаннной при компиляции
	buildVersion string
	buildDate    string
	buildCommit  string
)

// NewApplication инициализирует приложение
func NewApplication() (*Application, error) {
	// Инициализируем логгер.
	logFile, err := logger.NewLogFile("log.log")
	if err != nil {
		return nil, err
	}
	logger := logger.InitializeLogger(&log.JSONFormatter{}, log.DebugLevel, logFile)
	// Инициализируем конфигурацию
	conf, err := config.InitializeConfigurer(logger)
	if err != nil {
		logger.Debug("Сonfiguration initialization error.\nGophKeeper Server is stopped.\n")
		return nil, err
	}
	repo, err := repository.New(logger)
	if err != nil {
		logger.Debug("Errors occurred while connecting to the database.\nGophKeeper Server is stopped.\n")
		return nil, err
	}
	// Загружаем модуль шифрования
	crypt := cryptoargon.New(logger)
	// Запускаем сетевого клиента
	clnt := transport.NewTransporter(conf.GetHTTPSAddress(), conf.GetGRPCAddress(), logger)
	// Получем генератор QR кодов
	qrgen := qrgenerator.New()
	// получаем даннные версии
	versionControl()
	// Создаем GophKeeperService
	gophKeeper, err := gophkeeper.NewService(
		gophkeeper.Versions{BuildVersion: buildVersion, BuildDate: buildDate, BuildCommit: buildCommit},
		conf,
		repo,
		crypt,
		clnt,
		qrgen,
		logger,
	)
	if err != nil {
		logger.Debugf("Error starting service GophKeeper.\nGophKeeper Server is stopped.\n")
		return nil, err
	}
	// проверяем наличие данных о версии указанных при сборке
	tui := ui.New(gophKeeper, logger)
	// Создаем Application и делаем связи
	app := &Application{TUI: tui, Config: conf, Logger: logger}

	return app, nil
}

// RunApplication запускает приложение и обеспечивает безопасное завершение работы после получения
// сигналов os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT
func (a *Application) RunApplication() (chan struct{}, context.CancelFunc, error) {
	// Создаем канал для ожидания сигнала от системы на окончание работы
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	idleConnsClosed := a.waitSysSignals(ctx)
	// Запускаем TUI
	go a.TUI.Run(cancel)

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
		// получили сигнал, запускаем процедуру graceful shutdown сервера
		a.TUI.Program.Quit()

		// сообщаем основному потоку и логгируем, что все сетевые соединения обработаны и закрыты
		a.Logger.Debug("GophKeeper client has shutdown in graceful mode.")
		close(idleConnsClosed)
	}()
	return idleConnsClosed
}
