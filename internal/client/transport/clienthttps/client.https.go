// Package clienthttps реализует https клиента
package clienthttps

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"gophkeeper/pkg/domain"
	"gophkeeper/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Client описывает структуру HTTPS клиента
type Client struct {
	*http.Client
	ServerAddress string
	Logger        *log.Logger
}

// New создает HTTPS клиента
func New(serverAddr string, logger *log.Logger) *Client {
	certPool := x509.NewCertPool()

	certData, err := os.ReadFile("cert.pem")
	if err != nil {
		log.Fatal(err)
	}

	if !certPool.AppendCertsFromPEM(certData) {
		log.Fatal("failed to append cert")
	}

	tlsConfig := &tls.Config{
		RootCAs: certPool, // теперь клиент доверяет твоему серверу
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{Transport: tr}

	return &Client{Client: client, ServerAddress: serverAddr, Logger: logger}
}

// Request делает сетевой запрос методом method, по адресу endpoint,
func (c *Client) request(method string, endpoint string, contentType string, body string) (*http.Request, error) {
	request, err := http.NewRequest(method, endpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	// в заголовках запроса указываем тип данных
	request.Header.Add("Content-Type", contentType)
	return request, nil
}

// requestDo делает запрос к серверу
func (c *Client) requestDo(request *http.Request) (domain.Response, error) {
	var resp domain.Response
	// отправляем запрос и получаем ответ
	response, err := c.Do(request)
	if err != nil {
		return resp, err
	}
	// закрываем body после чтения ответа
	defer func() {
		err := response.Body.Close()
		if err != nil {
			c.Logger.Debug(err)
		}
	}()
	// читаем поток из тела ответа
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return resp, err
	}
	resp.Cookies = response.Cookies()
	resp.StatusCode = response.StatusCode
	resp.JSONData = string(result)

	return resp, nil
}

// Post делает POST-запрос к адресной строке endpoint с телом запроса body и типом контента contentType.
// В реализации клиент сервер используется только "application/json" в качестве content-type
func (c Client) newPost(endpoint string, body string) (*http.Request, error) {
	return c.request(http.MethodPost, "https://"+c.ServerAddress+endpoint, "application/json", body)
}

// Get делает GET-запрос к адресной строке endpoint и типом контента contentType.
// В реализации клиент сервер используется только "application/json" в качестве content-type
func (c Client) newGet(endpoint string) (*http.Request, error) {
	return c.request(http.MethodGet, "https://"+c.ServerAddress+endpoint, "application/json", "")
}

// Registration отправляет запрос регистрации к серверу
func (c Client) Registration(user *domain.User) (textErr string, err error) {
	request := domain.JSONRegistrationRequest{Login: user.Login, Password: user.Password, Email: user.Email}
	js, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	req, err := c.newPost("/api/user/registration", string(js))
	if err != nil {
		return "", err
	}
	resp, err := c.requestDo(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusConflict {
		return "", errors.ErrStatusConflict
	}

	var result domain.JSONRegistrationResponse
	if resp.JSONData != "" {
		if err := json.Unmarshal([]byte(resp.JSONData), &result); err != nil {
			return "", err
		}
	}
	user.TOTP = result.TOTPurl
	user.Session = result.Token
	user.SaltB64 = result.SaltB64

	return result.Error, nil
}

// Authorization отправляет запрос авторизации к серверу
func (c Client) Authorization(user *domain.User) (textErr string, err error) {
	request := domain.JSONAuthorizationRequest{Login: user.Login, Password: user.Password}
	js, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	req, err := c.newPost("/api/user/authorization", string(js))
	if err != nil {
		return "", err
	}
	resp, err := c.requestDo(req)
	if err != nil {
		return "", err
	}
	var result domain.JSONAuthorizationResponse
	if resp.JSONData != "" {
		if err := json.Unmarshal([]byte(resp.JSONData), &result); err != nil {
			return "", err
		}
	}
	user.Session = result.Token
	if result.TOTPUpdate {
		err = errors.ErrTOTPExpired
	}
	if result.SPassHash == "" {
		err = errors.ErrUserSecretPassword
	}
	user.SaltB64 = result.SaltB64
	user.SPassHash = result.SPassHash

	return result.Error, nil
}

// TOTPConfirmation отправляет запрос подтверждения одноразового кода к серверу
func (c Client) TOTPConfirmation(user *domain.User, code string) (textErr string, err error) {
	request := domain.JSONTOTPConfirmRequest{Code: code, Token: user.Session}
	js, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	req, err := c.newPost("/api/user/totp", string(js))
	if err != nil {
		return "", err
	}
	req.Header.Add("token", user.Session)
	resp, err := c.requestDo(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusForbidden {
		return "", errors.ErrStatusForbidden
	}
	var result domain.JSONTOTPConfirmResponse
	if resp.JSONData != "" {
		if err := json.Unmarshal([]byte(resp.JSONData), &result); err != nil {
			return "", err
		}
	}
	user.Token = result.AllowedToken

	return result.Error, nil
}

// TOTPUpdate отправляет на обновление данных TOTP авторизации
func (c Client) TOTPUpdate(user *domain.User, code string) (textErr string, err error) {
	request := domain.JSONTOTPUpdateRequest{Email: user.Email, Code: code, Token: user.Session}
	js, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	req, err := c.newPost("/api/user/totp-update", string(js))
	if err != nil {
		return "", err
	}
	req.Header.Add("token", user.Session)
	resp, err := c.requestDo(req)
	if err != nil {
		return "", err
	}
	var result domain.JSONTOTPUpdateResponse
	if resp.JSONData != "" {
		if err := json.Unmarshal([]byte(resp.JSONData), &result); err != nil {
			return "", err
		}
	}
	user.TOTP = result.TOTPurl

	return result.Error, nil
}

// SaveHashSecretPassword отправляет хэш пароля шифрования на сервер
func (c Client) SaveHashSecretPassword(user domain.User, hash string) (textErr string, err error) {
	request := domain.JSONSaveHashSecretPasswordRequest{AllowedToken: user.Token, Hash: hash}
	js, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	req, err := c.newPost("/api/user/save-hash", string(js))
	if err != nil {
		return "", err
	}
	req.Header.Add("token-allowed", user.Token)
	resp, err := c.requestDo(req)
	if err != nil {
		return "", err
	}
	var result domain.JSONErrorResponse
	if resp.JSONData != "" {
		if err := json.Unmarshal([]byte(resp.JSONData), &result); err != nil {
			return "", err
		}
	}

	return result.Error, nil
}

// SyncData синхронизирует данные с сервером
func (c Client) SyncData(dataOut domain.SyncPayload) (domain.SyncPayload, string, error) {
	js, err := json.Marshal(dataOut)
	if err != nil {
		return domain.SyncPayload{}, "", err
	}
	req, err := c.newPost("/api/sync", string(js))
	if err != nil {
		return domain.SyncPayload{}, "", err
	}
	req.Header.Add("token-allowed", dataOut.AllowedToken)
	resp, err := c.requestDo(req)
	if err != nil {
		return domain.SyncPayload{}, "", err
	}
	var dataIn domain.SyncPayload
	if resp.JSONData != "" {
		if err := json.Unmarshal([]byte(resp.JSONData), &dataIn); err != nil {
			return domain.SyncPayload{}, "", err
		}
	}

	return dataIn, dataIn.Error, nil
}
