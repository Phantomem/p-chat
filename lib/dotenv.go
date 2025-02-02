package lib

import (
	"os"
	"strconv"
)

var envCache = make(map[string][]byte)

func GetDotEnv(key string) []byte {
	if value, exists := envCache[key]; exists {
		return value
	}
	value := []byte(os.Getenv(key))
	if value == nil {
		panic("Error loading .env file")
	}
	envCache[key] = value
	return value
}

func GetNumDotEnv(key string) int {
	env, err := strconv.Atoi(string(GetDotEnv(key)))
	if err != nil {
		panic("Invalid ACCESS_TOKEN_SESSION_MINUTES value")
	}
	return env
}
