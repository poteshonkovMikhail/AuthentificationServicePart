package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	// Свои пакеты
	"auth_service_part/protobuf/protobuf_generated/auth_protobuf"
)

var upgrader = websocket.Upgrader{}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Не удалось открыть websocket соединение", http.StatusBadRequest)
		return
	}
	go handleTokenConnection(conn)
}

func handleTokenConnection(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка при попытке получения параметров запроса: %v", err)
			break
		}

		var req auth_protobuf.TokenRequest
		err = proto.Unmarshal(message, &req)
		if err != nil {
			log.Printf("Ошибка при демаршалировании параметров запроса: %v", err)
			continue
		}

		clientIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		ip := net.ParseIP(clientIP)

		accessToken, err := generateAccessToken(req.UserGuid, ip)
		if err != nil {
			log.Printf("Ошибка при попытке генерации Access-токена: %v", err)
			continue
		}

		refreshToken, err := generateRefreshToken()
		if err != nil {
			log.Printf("Ошибка при попытке генерации Refresh-токена: %v", err)
			continue
		}

		err = storeRefreshToken(req.UserGuid, refreshToken, ip)
		if err != nil {
			log.Printf("Ошибка при попытке хэширования или отправки в БД Refresh-токена: %v", err)
			continue
		}

		resp := &auth_protobuf.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		respData, _ := proto.Marshal(resp)
		conn.WriteMessage(websocket.BinaryMessage, respData)
	}
}
