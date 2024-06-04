package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    Port string
}

func LoadConfig() Config {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    return Config{
        Port: os.Getenv("CLIENT_SERVICE_PORT"),
    }
}
