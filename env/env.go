package env

import "os"

func GetEnvOrDefault(envVar, fallback string) string {
	value, exists := os.LookupEnv(envVar)
	if !exists {
		return fallback
	}
	return value
}
