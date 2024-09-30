package main

import (
	"fmt"
	"log"
	"time"

	"auth_service_part/protobuf/protobuf_generated/auth_protobuf"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto" // Убедитесь, что эта библиотека доступна
)

func main() {
	// Устанавливаем соединение с WebSocket
	url := "ws://localhost:8080/auth/token"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Ошибка при подключении к WebSocket:", err)
	}
	defer conn.Close()

	fmt.Println("Соединение открыто")

	// Создаем сообщение
	msg := &auth_protobuf.TokenRequest{
		UserGuid: "123e4567-e89b-12d3-a456-426614174000",
	}

	// Сериализуем сообщение в protobuf
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Println("Ошибка при сериализации сообщения:", err)
		return
	}

	// Отправляем сообщение на сервер
	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
		return
	}

	// Запускаем горутину для получения сообщений от сервера
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Ошибка при получении сообщения:", err)
				break
			}
			fmt.Println("Получено сообщение:", string(msg))
		}
	}()

	// Ждем некоторое время, прежде чем закрыть соединение
	time.Sleep(10 * time.Second)

	// Закрываем соединение
	fmt.Println("Закрываем соединение")
	err = conn.Close()
	if err != nil {
		log.Println("Ошибка при закрытии соединения:", err)
	}
}
