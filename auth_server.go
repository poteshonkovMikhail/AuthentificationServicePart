package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/peer"

	// Свои пакеты
	"auth_service_part/email_warning"
	"auth_service_part/helpers"
	"auth_service_part/protobuf/protobuf_generated/auth_protobuf"
)

type AuthServer struct {
	auth_protobuf.UnimplementedAuthServiceServer
}

var refreshTokenRotation_A_R = helpers.GetEnvAsBool("REFRESH_TOKEN_ROTATION_A_R", false)

func (s *AuthServer) GetToken(ctx context.Context, req *auth_protobuf.TokenRequest) (*auth_protobuf.TokenResponse, error) {
	userGuid := req.UserGuid

	// Получаем IP адрес клиента из контекста
	clientIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		clientIP, _, _ = net.SplitHostPort(p.Addr.String())
	}
	ip := net.ParseIP(clientIP)

	accessToken, err := generateAccessToken(userGuid, ip)
	if err != nil {
		log.Printf("Ошибка при попытке генерации Access-токена: %v", err)
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		log.Printf("Ошибка при попытке генерации Refresh-токена: %v", err)
		return nil, err
	}

	err = storeRefreshToken(userGuid, refreshToken, ip)
	if err != nil {
		log.Printf("Ошибка при попытке хэширования или отправки в БД Refresh-токена: %v", err)
		return nil, err
	}

	resp := &auth_protobuf.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}

func (s *AuthServer) OperationThatRefreshTokens(ctx context.Context, req *auth_protobuf.RefreshRequest) (*auth_protobuf.RefreshResponse, error) {
	token, err := jwt.Parse(req.AccessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("Medods_Secret_Key"), nil
	})

	if err != nil || !token.Valid {
		log.Printf("Недействительный Access-токен")
		return nil, fmt.Errorf("недействительный Access-токен: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Недействительные данные токена")
		return nil, fmt.Errorf("недействительные данные токена: %v", err)
	}

	userGUID := claims["user_guid"].(string)
	currentIP := net.ParseIP(claims["client_ip"].(string))

	isValid, storedIP, err := validateRefreshToken(userGUID, req.RefreshToken)
	if err != nil || !isValid {
		log.Printf("Недействительный Refresh-токен")
		return nil, fmt.Errorf("недействительный Refresh-токен: %v", err)
	}

	ipChanged := !currentIP.Equal(storedIP)

	if ipChanged {
		email_warning.SendEmailWarning(userGUID)
	}

	accessToken, err := generateAccessToken(userGUID, currentIP)
	if err != nil {
		log.Printf("Ошибка при попытке сгенерировать новый Access-токен: %v", err)
		return nil, err
	}

	if !refreshTokenRotation_A_R {
		resp := &auth_protobuf.RefreshResponse{
			AccessToken:  accessToken,
			RefreshToken: req.RefreshToken,
			IpChanged:    ipChanged,
		}
		return resp, nil
	} else {
		refreshToken, err := generateRefreshToken()
		if err != nil {
			log.Printf("Ошибка при попытке сгенерировать новый Refresh-токен: %v", err)
			return nil, err
		}

		err = storeRefreshToken(userGUID, refreshToken, currentIP)
		if err != nil {
			log.Printf("Ошибка при хэшировании или сохранении в БД нового Refresh-токена: %v", err)
			return nil, err
		}

		resp := &auth_protobuf.RefreshResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			IpChanged:    ipChanged,
		}
		return resp, nil
	}
}
