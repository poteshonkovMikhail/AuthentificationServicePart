package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

func initDB() {
	var err error
	// Определяем время ожидания для попыток подключения
	timeout := time.Minute
	startTime := time.Now()

	for {
		// Попытка установить соединение с postgresql сервером
		db, err = pgx.Connect(context.Background(), os.Getenv("POSTGRES_CONN"))
		if err == nil {
			// Проверка состояния подключения
			if err = db.Ping(context.Background()); err != nil {
				log.Fatalf("База данных недоступна: %v\n", err)
			} else {
				fmt.Println("Подключение к базе данных успешно")
				return
			}
		}

		// Проверяем, не истекло ли время ожидания
		if time.Since(startTime) > timeout {
			log.Fatalf("Не удалось установить соединение с PostgreSQL сервером: %v\n", err)
		}

		// Задержка перед следующей попыткой подключения
		time.Sleep(2 * time.Second) // Задержка 2 секунды
	}
}

// Инициализация маршрутов и запуск сервера
func main() {
	initDB()
	defer db.Close(context.Background())

	// Подумать над этим ↓
	http.HandleFunc("/auth/token", tokenHandler)
	http.HandleFunc("/auth/refresh", refreshHandler)

	if err := http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Println("Сервер запущен на порту :8080")
}
