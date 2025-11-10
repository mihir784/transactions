package config

import "os"

type Config struct {
	Port string
	PostgresURL string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{
		Port: port,
		PostgresURL: os.Getenv("POSTGRES_URL"),
	}
}
