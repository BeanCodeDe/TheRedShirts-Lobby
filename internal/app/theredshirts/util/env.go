package util

import (
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
)

func GetEnvWithFallback(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("environment variable %s not found", key)
}

func GetEnvIntWithFallback(key string, fallback int) (int, error) {
	if value, ok := os.LookupEnv(key); ok {
		return strconv.Atoi(value)
	}
	return fallback, nil
}

func GetEnvUUID(key string) (uuid.UUID, error) {
	if value, ok := os.LookupEnv(key); ok {
		return uuid.Parse(value)
	}
	return uuid.Nil, fmt.Errorf("environment variable %s not found", key)
}
