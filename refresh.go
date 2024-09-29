package main

import (
	"context"
	"encoding/base64"
	"math/rand"
	"net"

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
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Подумать над этим ↓
	_, err = db.Exec(
		context.Background(),
		"INSERT INTO refresh_tokens(user_guid, token_hash, client_ip) VALUES($1, $2, $3)",
		userGUID, hashedToken, clientIP.String(),
	)
	return err
}

// Валидация Refresh-токена
func validateRefreshToken(userGUID string, refreshToken string) (bool, net.IP, error) {
	var tokenHash string
	var clientIPStr string

	// Подумать над этим ↓
	err := db.QueryRow(
		context.Background(),
		"SELECT token_hash, client_ip FROM refresh_tokens WHERE user_guid=$1",
		userGUID,
	).Scan(&tokenHash, &clientIPStr)
	if err != nil {
		return false, nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(tokenHash), []byte(refreshToken))
	if err != nil {
		return false, nil, nil
	}

	clientIP := net.ParseIP(clientIPStr)
	return true, clientIP, nil
}
