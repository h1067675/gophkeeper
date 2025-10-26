// Package authorization реализует функции авторизации пользователя по сессиям и токенам, а также функцию двухфакторной авторизации
package authorization

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateSession генерирует слуайный код сессии для авторизации пользователей
func (a *Autorization) GenerateSession(allowed bool) (string, time.Time) {
	sessionDuration := 100 * 365 * 24 * time.Hour
	if allowed {
		sessionDuration = 12 * time.Hour
	}
	expire := time.Now().Add(sessionDuration)
	timestamp := expire.UnixNano()
	random := make([]byte, 8)
	_, _ = rand.Read(random)

	return fmt.Sprintf("%x-%x", timestamp, random), expire
}
