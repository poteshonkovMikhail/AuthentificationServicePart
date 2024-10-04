package main

import (
	"net"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Генерация JWT Access-токена
func generateAccessToken(userGUID string, clientIP net.IP) (string, error) {
	claims := jwt.MapClaims{
		"user_guid": userGUID,
		"client_ip": clientIP.String(),
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // Здесь SHA15
	return token.SignedString([]byte("Medods_Secret_Key"))
}
