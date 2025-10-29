// Package serverhttps реализует сервер HTTPS
package serverhttps

import (
	"compress/gzip"
	"context"
	"gophkeeper/pkg/domain"
	"io"
	"net/http"
	"strings"
	"time"
)

// структуры
type (

	// responseData описывает формат данных логгера
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter обертка над http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	// compressWriter описывает структуру необходимую для сжатия данных.
	compressWriter struct {
		http.ResponseWriter
		Writer io.Writer
	}
)

// Write реализует перехват метода http.ResponseWriter
func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.responseData.size += size
	return size, err
}

// WriteHeader реализует перехват метода http.ResponseWriter
func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.status = statusCode
}

// LoggingMiddleware логгирует данные запроса пользователя
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responce http.ResponseWriter, request *http.Request) {
		start := time.Now()
		responseData := responseData{
			status: 0,
			size:   0,
		}
		newResp := loggingResponseWriter{
			ResponseWriter: responce,
			responseData:   &responseData,
		}

		next.ServeHTTP(&newResp, request)
		s.Logger.WithFields(map[string]interface{}{
			"method":         request.Method,
			"URL":            request.URL.Path,
			"status":         newResp.responseData.status,
			"cookies":        request.Cookies(),
			"execution time": time.Since(start),
			"size":           newResp.responseData.size,
		}).Info("User request")
	})
}

// AuthorizationTokenMiddleware  осуществляет авторизацию пользователя прошедшего авторизацию по паролю и получившего код постоянной сессии.
func (s *Server) AuthorizationTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var ctx context.Context
		var session string
		if strings.Contains(request.RequestURI, domain.RouteUserRegistration) || strings.Contains(request.RequestURI, domain.RouteUserAuthorization) {
			next.ServeHTTP(response, request)
			return
		}
		// получаем токен пользователя
		token, ctxName := request.Header.Get("token-allowed"), sessionAllowed
		if token == "" {
			token, ctxName = request.Header.Get("token"), sessionID
			if token == "" {
				response.WriteHeader(http.StatusForbidden)
				return
			}
		}
		// забираем id сессии из токена
		session, err := s.GophKeeper.AuthorizationToken(token)
		if err != nil {
			response.WriteHeader(http.StatusForbidden)
			return
		}
		// передаем id сессии в хандлер
		ctx = context.WithValue(request.Context(), ctxName, session)

		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

// Write реализует метод интерфейса
func (c compressWriter) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

// CompressMiddleware промежуточный хэндлер отвечающий за сжатие данных
func (s *Server) CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
			if strings.Contains(request.Header.Get("Content-type"), "application/json") {
				gz, err := gzip.NewWriterLevel(response, gzip.BestSpeed)
				if err != nil {
					s.Logger.Debug(err)
					response.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer gz.Close()
				response.Header().Set("Content-Encoding", "gzip")
				response = compressWriter{ResponseWriter: response, Writer: gz}
			} else {
				response.WriteHeader(http.StatusUnsupportedMediaType)
			}
		}
		if strings.Contains(request.Header.Get("Content-Encoding"), "gzip") {
			cr, err := gzip.NewReader(request.Body)
			if err != nil {
				s.Logger.Debug(err)
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			request.Body = cr
			defer cr.Close()
		}
		next.ServeHTTP(response, request)
	})
}
