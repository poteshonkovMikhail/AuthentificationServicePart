package main

import (
	"context"
	"log"
	"net/http"
)

// Инициализация маршрутов и запуск сервера
func main() {
	InitDB()
	defer db.Close(context.Background())

	http.HandleFunc("/auth/token", someHandlerFunc)
	http.HandleFunc("/auth/refresh", someHandlerFunc)

	log.Println("Сервер запущен на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
