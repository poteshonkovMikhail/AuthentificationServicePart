package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	// Свои пакеты
	"auth_service_part/email_warning"
	"auth_service_part/helpers"
	"auth_service_part/protobuf/protobuf_generated/auth_protobuf"
)

var refreshTokenRotation_A_R = helpers.GetEnvAsBool("REFRESH_TOKEN_ROTATION_A_R", false)

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Не удалось открыть websocket соединение", http.StatusBadRequest)
		return
	}
	go handleRefreshConnection(conn)
}

func handleRefreshConnection(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка при попытке получения параметров запроса: %v", err)
			break
		}

		var req auth_protobuf.RefreshRequest
		err = proto.Unmarshal(message, &req)
		if err != nil {
			log.Printf("Ошибка при демаршалировании параметров запроса: %v", err)
			continue
		}

		token, err := jwt.Parse(req.AccessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte("Medods_Secret_Key"), nil
		})

		if err != nil || !token.Valid {
			log.Printf("Invalid Access-token")
			continue
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("Invalid token claims")
			continue
		}

		userGUID := claims["user_guid"].(string)               // Получение GUID пользователя из Payload Access-токена
		currentIP := net.ParseIP(claims["client_ip"].(string)) // Получение IP-адреса клиента из Payload Access-токена

		isValid, storedIP, err := validateRefreshToken(userGUID, req.RefreshToken)
		if err != nil || !isValid {
			log.Printf("Invalid Refresh-token")
			continue
		}

		ipChanged := !currentIP.Equal(storedIP) // Сравнение текущего IP клиента с сохраненным при прошлой сессии

		if ipChanged {
			email_warning.SendEmailWarning(userGUID) // Если IP изменился - отправляем email warning
		}

		accessToken, err := generateAccessToken(userGUID, currentIP) // Генерация нового Access-токена
		if err != nil {
			log.Printf("Ошибка при попытке сгенерировать новый Access-токен: %v", err)
			continue
		}

		// Возможность устанавливать тип ротации Refresh-токена через переменную окружения REFRESH_TOKEN_ROTATION_A_R
		// если FALSE --- Выдается новый только Access-токен, Refresh-токен остается прежним
		// если TRUE --- Выдаются новые Refresh- и Access- токены
		if !refreshTokenRotation_A_R {

			resp := &auth_protobuf.RefreshResponse{
				AccessToken:  accessToken,
				RefreshToken: req.RefreshToken,
				IpChanged:    ipChanged,
			}
			respData, _ := proto.Marshal(resp)
			conn.WriteMessage(websocket.BinaryMessage, respData)

		} else {

			refreshToken, err := generateRefreshToken() // Генерация нового Refresh-токена
			if err != nil {
				log.Printf("Ошибка при попытке сгенерировать новый Refresh-токен: %v", err)
				continue
			}

			err = storeRefreshToken(userGUID, refreshToken, currentIP) //Хэширование ---> Сохранение в БД Refresh-токена
			if err != nil {
				log.Printf("Ошибка при хэшировании или сохранении в БД нового Refresh-токена: %v", err)
				continue
			}

			resp := &auth_protobuf.RefreshResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				IpChanged:    ipChanged,
			}
			respData, _ := proto.Marshal(resp)
			conn.WriteMessage(websocket.BinaryMessage, respData)

		}
	}
}
