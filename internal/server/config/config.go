// Package config реализует первоначальнуюю загрузку конфигурации приложения
package config

import (
	"errors"
	"flag"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/caarlos0/env/v6"
)

// структуры
type (

	// Config хранит все необходимые настройки
	Config struct {
		Domain        StringEnv
		HTTPSPort     StringEnv
		GRPCPort      StringEnv
		DatabaseDNS   StringEnv
		SecretKey     StringEnv
		CompanyName   StringEnv
		SMTPHost      StringEnv
		SMTPPort      StringEnv
		Email         StringEnv
		EmailPassword StringEnv
		Logger        *log.Logger
	}

	// EnvConfig описывает формат парсинга переменных окружения
	EnvConfig struct {
		Domain        string `env:"GK_SERVER_DOMAIN"`
		HTTPSPort     string `env:"GK_HTTPS_PORT"`
		GRPCPort      string `env:"GK_GRPC_PORT"`
		DatabaseDNS   string `env:"GK_DATABASE_URI"`
		CompanyName   string `env:"GK_SMTPHost"`
		SMTPHost      string `env:"GK_SMTP_HOST"`
		SMTPPort      string `env:"GK_SMTP_PORT"`
		Email         string `env:"GK_EMAIL"`
		EmailPassword string `env:"GK_EMAIL_PASSWORD"`
		SecretKey     string `env:"GK_SECRET_KEY"`
	}

	// StringEnv описывает структуру необходимую для парсинга флагов приложения
	StringEnv struct {
		string
	}
)

// Set сохраняет значение текстовой переменной среды
func (n *StringEnv) Set(s string) error {
	n.string = s
	return nil
}

// String возвращает значение текстовой переменной
func (n *StringEnv) String() string {
	return n.string
}

// Check проверяет установлен ли параметр
func (n *StringEnv) Check(name string) error {
	if n.String() == "" {
		return fmt.Errorf("%s parameter is not filled in", name)
	}
	return nil
}

// InitializeConfigurer собирает конфигурацию сервера в следующем приоритете:
// 1. Флаги атрибутов командной строки
// 2. Переменные окружения
// 3. Параметры по умолчанию
//
// Внимание!!!
// Секретный ключ необходимый для шифрования данных указывается только в переменной среды GK_SECRET_KEY
// Пароль для отправки почтовых писем указывается только в переменной среды GK_EMAIL_PASSWORD
func InitializeConfigurer(domain, httpsport, grpcport, db, company, smtphost, smtpport, email string, logger *log.Logger) (conf *Config, err error) {
	conf = &Config{Logger: logger}
	err = errors.Join(
		conf.Domain.Set(domain),
		conf.HTTPSPort.Set(httpsport),
		conf.GRPCPort.Set(grpcport),
		conf.DatabaseDNS.Set(db),
		conf.CompanyName.Set(company),
		conf.SMTPHost.Set(smtphost),
		conf.SMTPPort.Set(smtpport),
		conf.Email.Set(email),
		conf.ParseEnvs(),
		conf.ParseFlags(),
	)
	if err != nil {
		conf.Logger.Debug(err)
	}
	if err := conf.CheckConfig(); err != nil {
		fmt.Printf("Сonfiguration is incorrect, please check settins and try again.")
		conf.PrintConfig()
		logger.Debug(err)
		return conf, err
	}
	if err != nil {
		fmt.Printf("Errors occurred during server configuration, please check the correctness of the configuration and confirm.")
		conf.PrintConfig()
		fmt.Printf("Is everything correct? (Y/n)")
		var y string
		fmt.Scan(&y)
		if y == "Y" {
			return conf, nil
		}
		return conf, err
	}
	return conf, nil
}

// ParseFlags разбирает атрибуты командной строки
func (c *Config) ParseFlags() error {
	flag.Var(&c.Domain, "a", "Host for runing servers")
	flag.Var(&c.HTTPSPort, "b", "Port for runing HTTPS server")
	flag.Var(&c.GRPCPort, "c", "Port for runing GRPC server")
	flag.Var(&c.DatabaseDNS, "d", "Data Source Name for accessing the database")
	flag.Var(&c.CompanyName, "n", "Company name")
	flag.Var(&c.SMTPHost, "sh", "SMTP email host")
	flag.Var(&c.SMTPPort, "sp", "SMTP email port")
	flag.Var(&c.Email, "e", "Email address for sending letters to users")
	flag.Parse()
	return nil
}

// ParseEnvs устанавливает настройки из переменных среды
func (c *Config) ParseEnvs() error {
	var envCnf EnvConfig
	err := env.Parse(&envCnf)
	if err != nil {
		c.Logger.WithError(err).Error("error parsing envs")
		return err
	}
	err = errors.Join(
		c.setParam(&c.Domain, envCnf.Domain),
		c.setParam(&c.HTTPSPort, envCnf.HTTPSPort),
		c.setParam(&c.GRPCPort, envCnf.GRPCPort),
		c.setParam(&c.DatabaseDNS, envCnf.DatabaseDNS),
		c.setParam(&c.CompanyName, envCnf.CompanyName),
		c.setParam(&c.SecretKey, envCnf.SecretKey),
		c.setParam(&c.SMTPHost, envCnf.SMTPHost),
		c.setParam(&c.SMTPPort, envCnf.SMTPPort),
		c.setParam(&c.Email, envCnf.Email),
		c.setParam(&c.EmailPassword, envCnf.EmailPassword),
	)
	return err
}

// setDefault заполняет параметр конфигурации значением
func (c *Config) setParam(value flag.Value, env string) error {
	if env != "" {
		err := value.Set(env)
		if err != nil {
			c.Logger.WithError(err).Debugf("error set configuration parameter from env")
			return err
		}
	}
	return nil
}

// CheckConfig проверяет все ли настройки были установлены
func (c *Config) CheckConfig() error {
	return errors.Join(
		checkNetAddress(fmt.Sprintf("%s:%s", c.Domain.String(), c.HTTPSPort.String())),
		checkNetAddress(fmt.Sprintf("%s:%s", c.Domain.String(), c.GRPCPort.String())),
		c.DatabaseDNS.Check("Database DNS"),
		c.SecretKey.Check("Secret key"),
	)
}

// PrintConfig выводит в консоль настройки сервера
func (c *Config) PrintConfig() {
	key := "empty"
	if c.SecretKey.String() != "" {
		key = "***setting***"
	}
	pass := "empty"
	if c.EmailPassword.String() != "" {
		pass = "***setting***"
	}
	fmt.Printf("Configuration:\n  Domain - %s\n  HTTPS port - %s\n  GRPC port - %s\n  Database DNS - %s\n  Company name - %s\n  SMTP host - %s\n  SMTP port - %s\n  Email address - %s\n  Secret key - %s\n  Email password - %s\n",
		c.Domain.String(),
		c.HTTPSPort.String(),
		c.GRPCPort.String(),
		c.DatabaseDNS.String(),
		c.CompanyName.String(),
		c.CompanyName.String(),
		c.CompanyName.String(),
		c.CompanyName.String(),
		key,
		pass,
	)
}

// checkNetAddress проверяет на корректность указания пары host:port
func checkNetAddress(str string) error {
	_, _, err := net.SplitHostPort(str)
	if err != nil {
		return err
	}
	return nil
}
