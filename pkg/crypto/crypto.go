// Package cryptoargon осуществляет шифрование данных
package cryptoargon

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"gophkeeper/pkg/errors"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
)

// CryptoArgon описывает настройки для шифрования данных
type CryptoArgon struct {
	ArgonTime    uint32
	ArgonMemory  uint32
	ArgonThreads uint8
	ArgonKeyLen  uint32
	SaltLength   int
	Logger       *logrus.Logger
}

// New создаети возвращает модуль криптозащищенного хэширования данных
func New(log *logrus.Logger) *CryptoArgon {
	return &CryptoArgon{
		ArgonTime:    2,
		ArgonMemory:  32 * 1024,
		ArgonThreads: 2,
		ArgonKeyLen:  32,
		SaltLength:   16,
		Logger:       log,
	}
}

// GenerateSalt создает соль
func (c *CryptoArgon) GenerateSalt() ([]byte, string, error) {
	salt := make([]byte, c.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, "", err
	}

	return salt, base64.RawStdEncoding.EncodeToString(salt), nil
}

// HashArgon2id возвращает PHC-строку хеша
func (c *CryptoArgon) HashArgon2id(data string) (string, error) {
	salt, b64Salt, err := c.GenerateSalt()
	if err != nil {
		return "", err
	}
	return c.Argon2id([]byte(data), salt, b64Salt)
}

// CompareHash проверяет соответствие пароля и PHC строки
func (c *CryptoArgon) CompareHash(data string, encodedHash string) (bool, error) {
	// проверяем количество параметров и соответствие формату: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.ErrInvalidHashFormat
	}
	// разбираем строку на параметры
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}
	var memory uint32
	var time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, err
	}
	// декодируем соль и хэш из base64
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	// вычисляем хеш от входных данных
	income := argon2.IDKey([]byte(data), salt, time, memory, threads, uint32(len(hash)))

	// безопасное сравнение
	if subtle.ConstantTimeCompare(hash, income) == 1 {
		return true, nil
	}
	return false, nil
}

// Encrypt осуществляет шифрование данных на клиенте используя секретный пароль пользователя
func (c *CryptoArgon) Encrypt(data []byte, password, salt64 string) (nonceAndCipherData string, err error) {
	// декодируем соль
	salt, err := base64.RawStdEncoding.DecodeString(salt64)
	if err != nil {
		return "", err
	}
	// делаем ключ шифрования уникальным
	key := argon2.IDKey([]byte(password), salt, c.ArgonTime, c.ArgonMemory, c.ArgonThreads, c.ArgonKeyLen)
	// AEAD
	aead, _ := chacha20poly1305.NewX(key)
	// генерируем уникальные данные для шифрования данных
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	rand.Read(nonce)
	// шифруем данные
	cipherData := aead.Seal(nil, nonce, data, nil)

	return fmt.Sprintf("%s|%s", base64.RawStdEncoding.EncodeToString(nonce), base64.RawStdEncoding.EncodeToString(cipherData)), nil
}

// Decrypt осуществляет расшифровку данных на клиенте используя секретный пароль пользователя
func (c *CryptoArgon) Decrypt(nonceAndCipherData, password, salt64 string) ([]byte, error) {
	parts := strings.Split(nonceAndCipherData, "|")
	nonce, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	data, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	// декодируем соль
	salt, err := base64.RawStdEncoding.DecodeString(salt64)
	if err != nil {
		return nil, err
	}
	// делаем ключ шифрования уникальным
	key := argon2.IDKey([]byte(password), salt, c.ArgonTime, c.ArgonMemory, c.ArgonThreads, c.ArgonKeyLen)
	// AEAD
	aead, _ := chacha20poly1305.NewX(key)
	return aead.Open(nil, nonce, data, nil)
}

// Argon2id возвращает PHC-строку хеша
func (c *CryptoArgon) Argon2id(data, salt []byte, b64Salt string) (encoded string, err error) {
	if salt == nil {
		salt, err = base64.RawStdEncoding.DecodeString(b64Salt)
		if err != nil {
			return "", err
		}
	}
	// вычисляем Argon2id
	hash := argon2.IDKey(data, salt, c.ArgonTime, c.ArgonMemory, c.ArgonThreads, c.ArgonKeyLen)
	// кодируем соль и хеш в base64 (URL-safe или Std)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	// соберём PHC-подобную строку
	encoded = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, c.ArgonMemory, c.ArgonTime, c.ArgonThreads, b64Salt, b64Hash)

	return encoded, nil
}
