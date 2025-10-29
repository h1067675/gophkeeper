// Package main GothKeeper Server
// терминальное приложение обеспечивающее безопасное хранение данных пользователя на сервере
//
// Сервер конфигурируется следующими аргументами:
// -a  :   Host for runing servers
// -b  :   Port for runing HTTPS server
// -c  :   Port for runing GRPC server
// -n  :   Company name
// -h  :   SMTP email host
// -sp :   SMTP email port
// -e  :   Email address for sending letters to users
//
// Секретный ключ необходимый для шифрования данных указывается только в переменной среды GK_SECRET_KEY
// Пароль для отправки почтовых писем указывается только в переменной среды GK_EMAIL_PASSWORD
//
// Остальные настройки также можно указать в переменных среды:
//
//	Domain           GK_SERVER_DOMAIN
//	HTTPSPort        GK_HTTPS_PORT
//	GRPCPort         GK_GRPC_PORT
//	DatabaseDNS  	 GK_DATABASE_URI
//	CompanyName   	 GK_SMTPHost
//	SMTPHost      	 GK_SMTP_HOST
//	SMTPPort         GK_SMTP_PORT
//	Email          	 GK_EMAIL
//	EmailPassword 	 GK_EMAIL_PASSWORD
//	SecretKey    	 GK_SECRET_KEY
package main

import (
	"fmt"

	"gophkeeper/internal/server/app"
)

func main() {
	dropDBTables := false // используется для тестирования и отладки во время разработки

	application, err := app.NewApplication(dropDBTables)
	if err != nil {
		fmt.Printf("Error initialize application.\nGophKeeper stopped.")
		return
	}
	idleConnsClosed, cancel, err := application.RunApplication()
	defer func() { cancel() }()
	if err != nil {
		fmt.Printf("Application launch error.\nGophKeeper stopped.")
		return
	}
	fmt.Printf("GophKeeper is running.\n")
	<-idleConnsClosed
	// Сообщаем об окончании работы программы
	fmt.Printf("GophKeeper has shutdown in graceful mode.\n")
}
