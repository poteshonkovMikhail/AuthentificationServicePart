package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Генерация Refresh-токена
func generateRefreshToken() (string, error) {
	byteToken := make([]byte, 64)
	_, err := rand.Read(byteToken)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(byteToken), nil
}

// Хэширование ---> Сохранение в БД Refresh-токена
func storeRefreshToken(userGUID string, refreshToken string, clientIP net.IP) error {
	// исправить
	//Проверка длины токена и обрезка, если необходимо
	if len(refreshToken) > 72 {
		refreshToken = refreshToken[:72]
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Ошибка хэширования токена: %v", err)
	}

	// Установка срока действия Refresh-токена
	expiresAt := time.Now().Add(2 * 24 * time.Hour)

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO refresh_tokens(user_guid, token_hash, client_ip, expires_at) VALUES($1, $2, $3, $4)",
		userGUID, hashedToken, clientIP.String(), expiresAt,
	)
	return fmt.Errorf("11111111111111 %v", err)
}

// Валидация Refresh-токена
func validateRefreshToken(userGUID string, refreshToken string) (bool, net.IP, error) {
	var tokenHash string
	var clientIPStr string
	var expiresAt time.Time

	err := db.QueryRow(
		context.Background(),
		"SELECT token_hash, client_ip, expires_at FROM refresh_tokens WHERE user_guid=$1",
		userGUID,
	).Scan(&tokenHash, &clientIPStr, &expiresAt)
	if err != nil {
		return false, nil, err
	}

	// Проверяем, не истек ли токен
	if time.Now().After(expiresAt) {
		return false, nil, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(tokenHash), []byte(refreshToken))
	if err != nil {
		return false, nil, nil
	}

	clientIP := net.ParseIP(clientIPStr)
	return true, clientIP, nil
}
