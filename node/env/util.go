package env

import "os"

func GetEnv(key string) string {
	if e := os.Getenv(key); e != "" {
		return e
	}
	if v, b := env[key]; b {
		return v
	}
	return ""
}
