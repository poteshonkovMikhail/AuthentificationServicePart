package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
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

// Шифрование данных
func encrypt(data string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Генерация Refresh-токена
func generateRefreshToken(userGUID string, clientIP net.IP) (string, error) {
	secret := []byte("Medods_Secret_Key_Tw2h3hdyr84hgr")

	randomBytes, err := generateRandomBytes(32)
	if err != nil {
		return "", err
	}

	encryptedData_1, err := encrypt(userGUID, secret)
	if err != nil {
		return "", err
	}

	encryptedData_2, err := encrypt(clientIP.String(), secret)
	if err != nil {
		return "", err
	}

	refreshToken := fmt.Sprintf("%s.%s.%s", base64.StdEncoding.EncodeToString(randomBytes), encryptedData_1, encryptedData_2)

	return refreshToken, nil
}

// Разделение строки Refresh-токена на части
func chunkString(s string, chunkSize int) []string {
	var chunks []string
	for len(s) > 0 {
		if len(s) < chunkSize {
			chunks = append(chunks, s)
			break
		}
		chunks = append(chunks, s[:chunkSize])
		s = s[chunkSize:]
	}
	return chunks
}

// Хэширование Refresh-токена с использованием соли
func hashToken(refreshToken string) (string, error) {
	chunks := strings.Split(refreshToken, ".")
	var hashedTokens []string

	for _, chunk := range chunks {

		hashedChunk, err := bcrypt.GenerateFromPassword([]byte(chunk), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		hashedTokens = append(hashedTokens, fmt.Sprintf("$%s", base64.StdEncoding.EncodeToString(hashedChunk)))
	}

	return strings.Join(hashedTokens, ""), nil
}

// Хэширование ---> Сохранение в БД Refresh-токена
func storeRefreshToken(userGUID string, refreshToken string, clientIP net.IP) error {
	hashedToken, err := hashToken(refreshToken)
	if err != nil {
		return fmt.Errorf("ошибка хэширования токена: %v", err)
	}

	// Установка срока действия Refresh-токена
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

	// Валидация каждого блока
	chunks := strings.Split(tokenHash, "$")
	refreshTokens := strings.Split(refreshToken, ".")

	log.Println(chunks)

	for i, chunk := range chunks {
		if chunk == "" {
			continue // Пропускаем пустые блоки
		}

		// Декодируем хэш
		hashedChunk, err := base64.StdEncoding.DecodeString(chunk) // Пропускаем символ '$'
		if err != nil {
			return false, nil, fmt.Errorf("ошибка декодирования хэша: %v", err)
		}

		// Сравниваем хэш с соответствующим блоком токена
		if err := bcrypt.CompareHashAndPassword(append([]byte("$"), hashedChunk...), []byte(refreshTokens[i])); err != nil {
			return false, nil, err // Токен недействителен
		}
	}

	clientIP := net.ParseIP(clientIPStr)
	return true, clientIP, nil
}

// Дешифрование данных AES
func decrypt(token string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:aes.BlockSize], ciphertext[aes.BlockSize:]

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Расшифровка Refresh-токена
func decodeRefreshToken(refreshToken string, key []byte) (string, net.IP, error) {
	parts := strings.Split(refreshToken, ".")
	if len(parts) != 2 {
		return "", nil, errors.New("invalid token format")
	}

	encryptedData := parts[1]

	decryptedData, err := decrypt(encryptedData, key)
	if err != nil {
		return "", nil, err
	}

	var data struct {
		UserGUID string `json:"user_guid"`
		ClientIP string `json:"client_ip"`
	}
	if err := json.Unmarshal([]byte(decryptedData), &data); err != nil {
		return "", nil, err
	}

	clientIP := net.ParseIP(data.ClientIP)
	if clientIP == nil {
		return "", nil, errors.New("invalid IP format")
	}

	return data.UserGUID, clientIP, nil
}
