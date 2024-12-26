package utils

import "os"

// GetEnv returns the value of an environment variable, or a fallback value if it is not set.
func GetEnv(key string, fallback ...string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	if len(fallback) > 0 {
		return fallback[0]
	}
	return ""
}
