package config

import(
	"os"
	"log"
	"github.com/joho/godotenv"
)

var Prefix string

func InitEnvronment() {
    if err := godotenv.Load(".env"); err != nil {           
        if err2 := godotenv.Load("../.env"); err2 != nil {  
            log.Fatalf("Error loading .env file : %v", err2)
        }
    }

    if os.Getenv("ENVIRONMENT") == "development" {
        Prefix = "_DEVELOPMENT"
    } else {
        Prefix = "_PRODUCTION"
    }
}


