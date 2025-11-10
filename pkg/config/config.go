package config

import "os"

func GetPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
