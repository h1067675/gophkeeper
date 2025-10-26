// Package qrgenerator генерирует QR коды
package qrgenerator

import (
	"github.com/skip2/go-qrcode"
)

// QRCode сущность модуля
type QRCode struct {
}

// New создает модуль
func New() *QRCode {
	return &QRCode{}
}

// QRMediumToString генерирует QR - код с данными из строки
func (q *QRCode) QRMediumToString(qrString string) (string, error) {
	qr, err := qrcode.New(qrString, qrcode.Medium)
	if err != nil {
		return "", err
	}

	return qr.ToSmallString(false), nil
}
