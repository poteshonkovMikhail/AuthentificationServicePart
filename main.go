package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

func initDB() {
	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("POSTGRES_CONN"))
	if err != nil {
		log.Fatalf("Не удалось установить соединение с PostgreSQL сервером: %v\n", err)
	}

	// Проверка состояния подключения к PostgreSQL серверу
	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("Database is unreachable: %v\n", err)
	} else {
		fmt.Println("Database connected successfully")
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
