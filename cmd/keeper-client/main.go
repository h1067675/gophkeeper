// GothKeeper Client
// терминальное приложение с TUI обеспечивающее безопасное хранение данных пользователя с авторизацией на сервере
// конфигурация приложения клиента производится из него
package main

import (
	"fmt"
	"gophkeeper/internal/client/app"
)

// main
func main() {
	application, err := app.NewApplication()
	if err != nil {
		fmt.Printf("Error initialize application.\nGophKeeper client stopped.")
		return
	}
	idleConnsClosed, cancel, err := application.RunApplication()
	defer func() { cancel() }()
	if err != nil {
		fmt.Printf("Application launch error.\nGophKeeper client stopped.")
		return
	}
	fmt.Printf("GophKeeper client is running.\n")
	<-idleConnsClosed
	// Сообщаем об окончании работы программы
	fmt.Printf("GophKeeper client has shutdown in graceful mode.\n")
}
