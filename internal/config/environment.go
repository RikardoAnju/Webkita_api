package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Prefix string

func InitEnvronment() {
	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load("../.env"); err2 != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	if os.Getenv("ENVIRONMENT") == "development" {
		Prefix = "_DEVELOPMENT"
	} else {
		Prefix = "_PRODUCTION"
	}
}