// Package serverhttps реализует сервер HTTPS
package serverhttps

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gophkeeper/pkg/domain"
	"gophkeeper/pkg/errors"
)

// GetErrorJSON формирует JSON с данными об ошибке
func GetErrorJSON(err error) ([]byte, error) {
	var js domain.JSONErrorResponse
	if err != nil {
		js.Error = err.Error()
	} else {
		js.Error = "unknown error"
	}
	res, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// getJSONData получает json данные из http.Request
func (s *Server) getJSONData(request *http.Request, vJSON any) (int, error) {
	// проверяем тип контента
	if !strings.Contains(request.Header.Get("Content-type"), "application/json") {
		return http.StatusUnsupportedMediaType, errors.ErrInvalidContentType
	}
	// читаем тело запроса
	var reqBody []byte
	reqBody, err := io.ReadAll(request.Body)
	if err != nil {
		s.Logger.Debug(err)
		return http.StatusInternalServerError, err
	}
	// читаем JSON данныые
	if err := json.Unmarshal(reqBody, vJSON); err != nil {
		s.Logger.Debug(err)
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, err
}

// UserRegistrationHandler фасад к функции UserRegistration.
// Получает на вход JSON следующего формата:
// login - логин пользователя
// password - пароль пользователя
// email - электронная почта пользователя
func (s *Server) UserRegistrationHandler(response http.ResponseWriter, request *http.Request) {
	var reqRegistration domain.JSONRegistrationRequest
	response.Header().Set("Content-Type", "application/json")
	statusCode, err := s.getJSONData(request, &reqRegistration)
	if err != nil {
		response.WriteHeader(statusCode)
		return
	}
	// если данные получены, то отправляем их в основную функцию
	token, expire, saltb64, totp, err := s.GophKeeper.UserRegistration(request.Context(), reqRegistration.Login, reqRegistration.Password, reqRegistration.Email)
	if err != nil {
		if errors.Is(err, errors.ErrUserAlreadyExist) {
			response.WriteHeader(http.StatusConflict)
			if body, err := GetErrorJSON(err); err == nil {
				response.Write(body)
			}
			return
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// формируем ответ в json и отправляем
	result := domain.JSONRegistrationResponse{Redirect: "/api/user/totp", TOTPurl: totp, Token: token, Expire: int(expire.Unix()), SaltB64: saltb64}
	body, err := json.Marshal(result)
	response.Write(body)
	response.WriteHeader(http.StatusOK)
}

// TOTPCheckHandler фасад к функции TOTPCheck
func (s *Server) TOTPCheckHandler(response http.ResponseWriter, request *http.Request) {
	var reqTOTPConfirmation domain.JSONTOTPConfirmRequest
	response.Header().Set("Content-Type", "application/json")
	statusCode, err := s.getJSONData(request, &reqTOTPConfirmation)
	if err != nil {
		response.WriteHeader(statusCode)
		return
	}
	ctx := request.Context()
	// если данные получены, то отправляем их в основную функцию
	token, expire, err := s.GophKeeper.ValidateTOTPCode(ctx, ctx.Value(sessionID).(string), reqTOTPConfirmation.Code)
	if err != nil {
		if errors.Is(err, errors.ErrTOTPCodeCheck) {
			response.WriteHeader(http.StatusConflict)
			if body, err := GetErrorJSON(err); err == nil {
				response.Write(body)
			}
			return
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	// если новая сессия создана без ошибок то в куки прописываем токен
	result := domain.JSONTOTPConfirmResponse{AllowedToken: token, Expire: int(expire.Unix())}
	body, err := json.Marshal(result)
	response.Write(body)
	// отправляем статус 200
	response.WriteHeader(http.StatusOK)
}

// TOTPUpdateHandler фасад к функции TOTPUpdate
func (s *Server) TOTPUpdateHandler(response http.ResponseWriter, request *http.Request) {
	var reqTOTPUpdate domain.JSONTOTPUpdateRequest
	response.Header().Set("Content-Type", "application/json")
	statusCode, err := s.getJSONData(request, &reqTOTPUpdate)
	if err != nil {
		response.WriteHeader(statusCode)
		return
	}
	ctx := request.Context()
	// если данные получены, то отправляем их в основную функцию
	totp, err := s.GophKeeper.UpdateTOTPRegistration(ctx, ctx.Value(sessionID).(string), reqTOTPUpdate.Email, reqTOTPUpdate.Code)
	if err != nil {
		if errors.Is(err, errors.ErrTOTPCodeCheck) {
			response.WriteHeader(http.StatusConflict)
			if body, err := GetErrorJSON(err); err == nil {
				response.Write(body)
			}
			return
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	// формируем ответ в json и отправляем
	result := domain.JSONRegistrationResponse{Redirect: "/api/user/totp", TOTPurl: totp}
	body, err := json.Marshal(result)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
	}
	response.Write(body)
	response.WriteHeader(http.StatusOK)
}

// UserAuthorizationHandler фасад к функции UserRegistration.
// Получает на вход JSON следующего формата:
// login - логин пользователя
// password - пароль пользователя
func (s *Server) UserAuthorizationHandler(response http.ResponseWriter, request *http.Request) {
	var reqAuthorization domain.JSONAuthorizationRequest
	response.Header().Set("Content-Type", "application/json")
	statusCode, err := s.getJSONData(request, &reqAuthorization)
	if err != nil {
		response.WriteHeader(statusCode)
		return
	}
	// если данные получены, то отправляем их в основную функцию
	session, expire, saltb64, spassHash, errAuth := s.GophKeeper.UserAuthorization(request.Context(), reqAuthorization.Login, reqAuthorization.Password)
	if errAuth != nil && !errors.Is(errAuth, errors.ErrTOTPExpired) && !errors.Is(errAuth, errors.ErrUserSecretPassword) {
		if errors.Is(errAuth, errors.ErrUserAuthorization) {
			response.WriteHeader(http.StatusForbidden)
			if body, err := GetErrorJSON(errAuth); err == nil {
				response.Write(body)
			}
			return
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	// формируем ответ в json
	result := domain.JSONAuthorizationResponse{Redirect: "/api/user/totp", Token: session, Expire: int(expire.Unix()), SaltB64: saltb64, SPassHash: spassHash}
	statusCode = http.StatusOK
	if errors.Is(errAuth, errors.ErrTOTPExpired) {
		result.Redirect = "/api/user/totp-update"
		statusCode = http.StatusTemporaryRedirect
	}
	if errors.Is(errAuth, errors.ErrUserSecretPassword) {
		result.Redirect = "/api/user/save-hash"
		statusCode = http.StatusTemporaryRedirect
	}

	body, err := json.Marshal(result)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.WriteHeader(statusCode)
	response.Write(body)
}

// UserSavePasswordHashHandler фасад к функции UserSavePasswordHash.
func (s *Server) UserSavePasswordHashHandler(response http.ResponseWriter, request *http.Request) {
	var reqSavePassHash domain.JSONSaveHashSecretPasswordRequest
	response.Header().Set("Content-Type", "application/json")
	statusCode, err := s.getJSONData(request, &reqSavePassHash)
	if err != nil {
		response.WriteHeader(statusCode)
		return
	}
	ctx := request.Context()
	// если данные получены, то отправляем их в основную функцию
	err = s.GophKeeper.UserSavePasswordHash(ctx, ctx.Value(sessionAllowed).(string), reqSavePassHash.Hash)
	if err != nil {
		if body, err := GetErrorJSON(err); err == nil {
			response.Write(body)
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.WriteHeader(statusCode)
}

// SyncDataHandler фасад к функции UserSavePasswordHash.
func (s *Server) SyncDataHandler(response http.ResponseWriter, request *http.Request) {
	var reqData domain.SyncPayload
	response.Header().Set("Content-Type", "application/json")
	statusCode, err := s.getJSONData(request, &reqData)
	if err != nil {
		response.WriteHeader(statusCode)
		return
	}
	ctx := request.Context()
	// если данные получены, то отправляем их в основную функцию
	resData, err := s.GophKeeper.SyncData(ctx, ctx.Value(sessionAllowed).(string), reqData)
	if err != nil {
		if body, err := GetErrorJSON(err); err == nil {
			response.Write(body)
		}
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	// формируем ответ в json и отправляем
	body, err := json.Marshal(resData)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
	}
	response.Write(body)

	response.WriteHeader(statusCode)
}
