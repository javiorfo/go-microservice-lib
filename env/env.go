package env

import (
	"encoding/json"
	"os"
)

func GetEnvOr[T any](envVar string, fallback T) T {
	value, exists := os.LookupEnv(envVar)
	if !exists {
		return fallback
	}

	var result T
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		return fallback
	}

	return result
}
