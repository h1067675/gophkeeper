// Package domain описывает форматы сущностей всего проекта
package domain

type (
	// JSONErrorResponse ответ сервера содержаний тольо ошибку
	JSONErrorResponse struct {
		Error string `json:"error"`
	}
	// JSONRegistrationRequest запрос на регистрацию
	JSONRegistrationRequest struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// JSONRegistrationResponse ответ на запрос о регистрации
	JSONRegistrationResponse struct {
		Token    string `json:"token"`
		SaltB64  string `json:"saltb64"`
		Expire   int    `json:"expire"`
		Redirect string `json:"location"`
		TOTPurl  string `json:"totp_url"`
		Error    string `json:"error"`
	}

	// JSONTOTPConfirmRequest запрос TOTP код
	JSONTOTPConfirmRequest struct {
		Token string `json:"token"`
		Code  string `json:"code"`
	}

	// JSONTOTPConfirmResponse ответ на запрос TOTP кода
	JSONTOTPConfirmResponse struct {
		AllowedToken string `json:"allowed_token"`
		Expire       int    `json:"expire"`
		Error        string `json:"error"`
	}

	// JSONTOTPUpdateRequest запрос на обновление TOTP
	JSONTOTPUpdateRequest struct {
		Token string `json:"token"`
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	// JSONTOTPUpdateResponse ответ на запрос на обновление TOTP
	JSONTOTPUpdateResponse struct {
		TOTPurl string `json:"totp_url"`
		Error   string `json:"error"`
	}

	// JSONAuthorizationRequest запрос на авторизацию
	JSONAuthorizationRequest struct {
		Login    string `json:"user_login"`
		Password string `json:"password_token"`
	}

	// JSONAuthorizationResponse ответ на запрос на авторизацию
	JSONAuthorizationResponse struct {
		Token      string `json:"token"`
		SaltB64    string `json:"saltb64"`
		SPassHash  string `json:"spass_hash"`
		Redirect   string `json:"location"`
		Expire     int    `json:"expire"`
		TOTPUpdate bool   `json:"totp_update"`
		Error      string `json:"error"`
	}

	// JSONSaveHashSecretPasswordRequest запрос на сохранение хэша пароля шифрования
	JSONSaveHashSecretPasswordRequest struct {
		AllowedToken string `json:"allowed_token"`
		Hash         string `json:"hash"`
	}
)
