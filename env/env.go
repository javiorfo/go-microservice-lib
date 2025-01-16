package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func GetEnvOr[T any](envVar string, fallback T) T {
    // Uses .env file if exists, if not, get the fallback
    _ = godotenv.Load()

	value, exists := os.LookupEnv(envVar)
	if !exists {
		return fallback
	}

	var result T
	switch any(result).(type) {
	case int:
		i, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		result = any(i).(T)
	case float64:
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fallback
		}
		result = any(i).(T)
	case bool:
		i, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		result = any(i).(T)
	case string, []byte:
		result = any(value).(T)
	default:
		return fallback
	}

	return result
}
