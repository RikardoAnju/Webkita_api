package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"BackendFramework/internal/config"
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	v1 "BackendFramework/internal/route/v1"
)

func init() {
	// Load ENV
	config.InitEnvronment()

	log.Println("🔧 Initializing middleware...")
	middleware.InitLogger()
	middleware.InitValidator()

	log.Println("🔧 Opening database connection...")
	database.Connect()

	log.Println("✅ Initialization complete!")
}

func main() {
	log.Println("🛠️ Running database migrations...")
	database.RunMigrations()
	defer database.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()

	// Routes
	v1.SetupRoutes(router, database.DbWebkita)

	log.Println("🚀 Starting server on port:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}