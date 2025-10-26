// Package config реализует первоначальнуюю загрузку конфигурации приложения
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"gophkeeper/pkg/errors"

	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
)

// структуры
type (

	// Config хранит все необходимые настройки
	Config struct {
		Domain    StringEnv
		HTTPSPort StringEnv
		GRPCPort  StringEnv
		Session   StringEnv
		SaltB64   StringEnv
		SPassHash StringEnv
		Ok        bool
		Logger    *log.Logger
	}

	// EnvConfig описывает формат парсинга переменных окружения
	EnvConfig struct {
		Domain    string `env:"GK_SERVER_DOMAIN"`
		HTTPSPort string `env:"GK_HTTPS_PORT"`
		GRPCPort  string `env:"GK_GRPC_PORT"`
	}

	// StringEnv описывает структуру необходимую для парсинга флагов приложения
	StringEnv struct {
		string
	}

	// JSONConfig описывает структуру парсинга JSON конфига
	JSONConfig struct {
		Domain    string `json:"server_domain"`
		HTTPSPort string `json:"https_port"`
		GRPCPort  string `json:"grpc_port"`
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

// saveToKeyringService сохраняет данные в хранилище паролей операционной системы
func (n *StringEnv) saveToKeyringService(appName, keyName, value string) error {
	if err := n.Set(value); err != nil {
		return err
	}
	return keyring.Set(appName, keyName, value)
}

// getFromKeyringService берет данные из хранилища паролей операционной системы
func (n *StringEnv) getFromKeyringService(appName, keyName string) error {
	v, err := keyring.Get(appName, keyName)
	if err != nil {
		return err
	}
	return n.Set(v)
}

// deleteFromKeyringService удаляет данные из хранилища паролей операционной системы
func (n *StringEnv) deleteFromKeyringService(appName, keyName string) error {
	return errors.Join(keyring.Delete(appName, keyName), n.Set(""))
}

// InitializeConfigurer собирает конфигурацию сервера из переменных окружения, JSON файла и хранилища паролей операционной системы
func InitializeConfigurer(logger *log.Logger) (conf *Config, err error) {
	conf = &Config{Logger: logger}
	err = errors.Join(
		conf.ParseEnvs(),
	)
	if err != nil {
		conf.Logger.Debug(err)
	}
	jscfg, err := conf.GetConfigFromJSONFile()
	if err != nil {
		conf.Logger.Debug(err)
	}
	err = errors.Join(conf.Domain.Set(jscfg.Domain), conf.GRPCPort.Set(jscfg.GRPCPort), conf.HTTPSPort.Set(jscfg.HTTPSPort))
	if err != nil {
		logger.Debug(err)
	}
	if err := conf.Session.getFromKeyringService("GophKeeper", "session"); err != nil {
		logger.Debug(err)
	}

	if err := conf.SaltB64.getFromKeyringService("GophKeeper", "saltb64"); err != nil {
		logger.Debug(err)
	}

	if err := conf.SPassHash.getFromKeyringService("GophKeeper", "spasshash"); err != nil {
		logger.Debug(err)
	}

	return conf, nil
}

// ParseEnvs устанавливает настройки из переменных среды
func (c *Config) ParseEnvs() error {
	var envCnf EnvConfig
	err := env.Parse(&envCnf)
	if err != nil {
		c.Logger.Debug(err)
		return err
	}
	err = errors.Join(
		c.setParam(&c.Domain, envCnf.Domain),
		c.setParam(&c.HTTPSPort, envCnf.HTTPSPort),
		c.setParam(&c.GRPCPort, envCnf.GRPCPort),
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

// CheckConfig проверяет установлен ли хотябы один метод передачи данных
func (c *Config) CheckConfig() error {
	if c.CheckNetAddress(fmt.Sprintf("%s:%s", c.Domain.String(), c.HTTPSPort.String())) == nil ||
		c.CheckNetAddress(fmt.Sprintf("%s:%s", c.Domain.String(), c.GRPCPort.String())) == nil {
		return nil
	}
	return errors.Join(
		c.CheckNetAddress(fmt.Sprintf("%s:%s", c.Domain.String(), c.HTTPSPort.String())),
		c.CheckNetAddress(fmt.Sprintf("%s:%s", c.Domain.String(), c.GRPCPort.String())),
	)
}

// CheckNetAddress проверяет на корректность указания пары host:port
func (c *Config) CheckNetAddress(str string) error {
	host, port, err := net.SplitHostPort(str)
	if err != nil {
		return err
	} else if host == "" || port == "" {
		return errors.ErrInvalideNetAddress
	}
	return nil
}

// GetServerAddress реализует интерфейсный метод для передачи адреса сервера
func (c *Config) GetServerAddress() (serverAddress string, err error) {
	return c.Domain.String(), nil
}

// GetGRPCAddress реализует интерфейсный метод для передачи адреса GRPC вида host:port
func (c *Config) GetGRPCAddress() (address string) {
	addr := fmt.Sprintf("%s:%s", c.Domain.String(), c.GRPCPort.String())
	if c.CheckNetAddress(addr) == nil {
		return addr
	}
	return ""
}

// GetHTTPSAddress реализует интерфейсный метод для передачи адреса HTTPS вида host:port
func (c *Config) GetHTTPSAddress() (address string) {
	addr := fmt.Sprintf("%s:%s", c.Domain.String(), c.HTTPSPort.String())
	if c.CheckNetAddress(addr) == nil {
		return addr
	}
	return ""
}

// SetConfig устанавливает настройки адресов указанные пользователем
func (c *Config) SetConfig(serverAddress, grpcPort, httpsPort string) error {
	return errors.Join(c.Domain.Set(serverAddress),
		c.GRPCPort.Set(grpcPort),
		c.HTTPSPort.Set(httpsPort), c.SaveToFile())
}

// GetConfigFromJSONFile получает конфигурацию из файла JSON
func (c *Config) GetConfigFromJSONFile() (JSONConfig, error) {
	var cfg JSONConfig
	var err error
	// получаем директорию текущего файла
	appfile, err := os.Executable()
	if err != nil {
		return cfg, err
	}
	// читаем файл конфигурации
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), "./settings.json"))
	if err != nil {
		return cfg, err
	}
	// помещаем настройки из файла в структуру
	if err = json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// SaveToFile сохраняет конфигурацию в файл JSON
func (с *Config) SaveToFile() error {
	cfg := JSONConfig{Domain: с.Domain.String(), HTTPSPort: с.HTTPSPort.String(), GRPCPort: с.GRPCPort.String()}
	js, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	appfile, err := os.Executable()
	if err != nil {
		return err
	}
	fl, err := os.Create(filepath.Join(filepath.Dir(appfile), "./settings.json"))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := fl.Close(); cerr != nil {
			с.Logger.Debug(err)
		}
	}()
	_, err = fl.Write(js)
	if err != nil {
		return err
	}
	return nil
}

// SaveUserSession интерфейсный метод сохраняющий данные сессии в хранилище паролей ОС
func (c *Config) SaveUserSession(token string) error {
	return c.Session.saveToKeyringService("GophKeeper", "session", token)
}

// DeleteUserSession интерфейсный метод удаляющий данные сессии из хранилища паролей ОС
func (c *Config) DeleteUserSession() error {
	return c.Session.deleteFromKeyringService("GophKeeper", "session")
}

// GetUserSession интерфейсный метод получающий данные сессии из хранилища паролей ОС
func (c *Config) GetUserSession() string {
	return c.Session.String()
}

// SaveUserSaltB64 интерфейсный метод сохраняющий шифровальную соль в хранилище паролей ОС
func (c *Config) SaveUserSaltB64(saltb64 string) error {
	return c.SaltB64.saveToKeyringService("GophKeeper", "saltb64", saltb64)
}

// DeleteUserSaltB64 интерфейсный метод удаляющий шифровальную соль из хранилища паролей ОС
func (c *Config) DeleteUserSaltB64() error {
	return c.SaltB64.deleteFromKeyringService("GophKeeper", "saltb64")
}

// GetUserSaltB64 интерфейсный метод получающий шифровальную соль из хранилища паролей ОС
func (c *Config) GetUserSaltB64() string {
	return c.SaltB64.String()
}

// SaveUserSPassHash интерфейсный метод сохраняющий хэш крипто пароля в хранилище паролей ОС
func (c *Config) SaveUserSPassHash(sPassHash string) error {
	return c.SPassHash.saveToKeyringService("GophKeeper", "spasshash", sPassHash)
}

// DeleteUserSPassHash интерфейсный метод удаляющий хэш крипто пароля из хранилища паролей ОС
func (c *Config) DeleteUserSPassHash() error {
	return c.SPassHash.deleteFromKeyringService("GophKeeper", "spasshash")
}

// GetUserSPassHash интерфейсный метод получающий хэш крипто пароля из хранилища паролей ОС
func (c *Config) GetUserSPassHash() string {
	return c.SPassHash.String()
}
