package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	SaltSize     = 72
	GUIDSize     = 72
	ClientIPSize = 72
)

// Генерация случайной строки
func generateRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Генерация Refresh-токена
func generateRefreshToken(userGUID string, clientIP net.IP) (string, error) {
	// Генерация случайной соли
	randomBytes, err := generateRandomBytes(SaltSize)
	if err != nil {
		return "", err
	}
	salt := base64.StdEncoding.EncodeToString(randomBytes)[:SaltSize]
	userGUID = base64.StdEncoding.EncodeToString([]byte(userGUID))
	clientIPEncode := base64.StdEncoding.EncodeToString([]byte(clientIP.String()))

	refreshToken := fmt.Sprintf("%s|%s|%s", salt, userGUID, clientIPEncode)
	log.Println(refreshToken)
	return refreshToken, nil
}

func hashToken(refreshToken string) (string, error) {
	chunks := strings.Split(refreshToken, "|")
	if len(chunks) != 3 {
		return "", errors.New("недопустимая структура токена")
	}

	var hashedTokens []string
	for _, chunk := range chunks {
		hashedChunk, err := bcrypt.GenerateFromPassword([]byte(chunk), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		hashedTokens = append(hashedTokens, string(hashedChunk)) // Изменим тут
	}

	return strings.Join(hashedTokens, "|"), nil // изменение тут
}

// Сохранение токена в БД
func storeRefreshToken(userGUID string, refreshToken string, clientIP net.IP) error {
	log.Println(refreshToken)

	hashedToken, err := hashToken(refreshToken)
	if err != nil {
		return fmt.Errorf("ошибка хэширования токена: %v", err)
	}

	expiresAt := time.Now().Add(2 * 24 * time.Hour)

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO refresh_tokens(user_guid, token_hash, client_ip, expires_at) VALUES($1, $2, $3, $4)",
		userGUID, hashedToken, clientIP.String(), expiresAt,
	)
	if err != nil {
		return fmt.Errorf("не удалось создать ваш токен: %v", err)
	}

	return nil
}

// Валидация токена
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

	if time.Now().After(expiresAt) {
		removeNonValidToken(userGUID)
		return false, nil, nil
	}

	updatedRefreshToken := strings.ReplaceAll(refreshToken, " ", "+")

	hashChunks := strings.Split(tokenHash, "|")
	tokenChunks := strings.Split(updatedRefreshToken, "|")
	log.Println(updatedRefreshToken)

	if len(hashChunks) != len(tokenChunks) {
		return false, nil, errors.New("invalid token structure")
	}

	for i, hashedChunk := range hashChunks {
		if err := bcrypt.CompareHashAndPassword([]byte(hashedChunk), []byte(tokenChunks[i])); err != nil {
			return false, nil, err
		}
	}

	clientIP := net.ParseIP(clientIPStr)
	return true, clientIP, nil
}

// Парсинг Payload'a Refresh-токена
func parseRefreshToken(refreshToken string) (string, net.IP, error) {
	chunks := strings.Split(refreshToken, "|")
	if len(chunks) != 3 {
		return "", nil, errors.New("недопустимая структура токена")
	}

	userGUIDBytes, err := base64.StdEncoding.DecodeString(chunks[1])
	if err != nil {
		return "", nil, fmt.Errorf("ошибка декодирования userGUID: %v", err)
	}
	userGUID := string(userGUIDBytes)
	log.Println(userGUID)
	clientIPBytes, err := base64.StdEncoding.DecodeString(chunks[2])
	if err != nil {
		return "", nil, fmt.Errorf("ошибка декодирования clientIP: %v", err)
	}

	clientIP := net.ParseIP(string(clientIPBytes))
	if clientIP == nil {
		return "", nil, fmt.Errorf("не удалось распознать IP-адрес из: %s", string(clientIPBytes))
	}

	return userGUID, clientIP, nil
}
