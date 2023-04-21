package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_HOST string
	DB_USER string
	DB_PASS string
	DB_PORT string
	DB_NAME string

	JWT_SECRET string
}

func InitConfig() *Config {
	godotenv.Load()

	return &Config{
		DB_HOST:    os.Getenv("DB_HOST"),
		DB_USER:    os.Getenv("DB_USER"),
		DB_PASS:    os.Getenv("DB_PASS"),
		DB_PORT:    os.Getenv("DB_PORT"),
		DB_NAME:    os.Getenv("DB_NAME"),
		JWT_SECRET: os.Getenv("JWT_SECRET"),
	}
}
