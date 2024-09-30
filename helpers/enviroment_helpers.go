package helpers

import (
	"os"
	"strconv"
)

// Хэлпер для чтения переменной окружения, как булевого значения (если переменной нет, то возвращает дефолтное значение)
func GetEnvAsBool(name string, defaultVal bool) bool {
	valStr := GetEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Хэлпер для чтения переменной окружения, как строки (если переменной нет, то возвращает дефолтное значение)
func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
