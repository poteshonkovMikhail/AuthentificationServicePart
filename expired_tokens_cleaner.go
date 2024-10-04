package main

import (
	"context"
	"log"
	"time"
)

func removeExpiredTokens() {
	for {
		now := time.Now()

		// Удаление истекших токенов
		_, err := db.Exec(context.Background(),
			"DELETE FROM refresh_tokens WHERE expires_at < $1", now)
		if err != nil {
			log.Printf("Ошибка при удалении истекших токенов: %v\n", err)
		} else {
			log.Println("Истекшие Refresh-токены успешно удалены")
		}

		// Интервал выполнения - 1 час
		time.Sleep(1 * time.Hour)
	}
}

func removeNonValidToken(userGuid string) {
	// Удаление истекших токенов
	_, err := db.Exec(context.Background(),
		"DELETE FROM refresh_tokens WHERE user_guid = $1", userGuid)
	if err != nil {
		log.Printf("Ошибка при удалении истекших токенов: %v\n", err)
	} else {
		log.Println("Истекшие Refresh-токены успешно удалены")
	}
}
