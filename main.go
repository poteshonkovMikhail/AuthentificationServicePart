package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

func initDB() {
	var err error
	db, err = pgx.Connect(context.Background(), "postgresql://user:password@localhost:5432/authdb")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
}

// Инициализация маршрутов и запуск сервера
func main() {
	initDB()
	defer db.Close(context.Background())

	// Подумать над этим ↓
	http.HandleFunc("/auth/token", tokenHandler)
	http.HandleFunc("/auth/refresh", someHandlerFunc)

	log.Println("Сервер запущен на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
