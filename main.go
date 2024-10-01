package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	// Свои пакеты
	"auth_service_part/protobuf/protobuf_generated/auth_protobuf"
)

var db *pgx.Conn

func initDB() {
	var err error
	timeout := time.Minute
	startTime := time.Now()

	for {
		db, err = pgx.Connect(context.Background(), os.Getenv("POSTGRES_CONN"))
		if err == nil {
			if err = db.Ping(context.Background()); err != nil {
				log.Fatalf("База данных недоступна: %v\n", err)
			} else {
				fmt.Println("Подключение к базе данных успешно")
				return
			}
		}

		if time.Since(startTime) > timeout {
			log.Fatalf("Не удалось установить соединение с PostgreSQL сервером: %v\n", err)
		}

		time.Sleep(2 * time.Second)
	}
}

func main() {
	initDB()
	defer db.Close(context.Background())

	// Создаем gRPC сервер
	grpcServer := grpc.NewServer()
	authService := &AuthServer{}
	auth_protobuf.RegisterAuthServiceServer(grpcServer, authService)

	// Включаем reflection для gRPC сервера (опционально)
	reflection.Register(grpcServer)

	// Запускаем gRPC сервер
	go func() {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			log.Fatalf("Не удалось начать прослушивание: %v", err)
		}
		log.Println("gRPC сервер запущен на порту :9090")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Не удалось запустить gRPC сервер: %v", err)
		}
	}()

	// Создаем gRPC-Gateway
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := auth_protobuf.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, ":9090", opts)
	if err != nil {
		log.Fatalf("Не удалось зарегистрировать обработчик gRPC-Gateway: %v", err)
	}

	log.Println("gRPC-Gateway запущен на порту :8080")
	if err := http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
