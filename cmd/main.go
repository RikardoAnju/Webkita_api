package main

import (
    "log"
    "os"

    "BackendFramework/internal/config"
    "BackendFramework/internal/database"
    "BackendFramework/internal/middleware"
    "BackendFramework/internal/route"
)

func init() {
    // Load .env dulu via InitEnvronment
    config.InitEnvronment()

    log.Println("🔧 Initializing middleware...")
    middleware.InitLogger()
    middleware.InitValidator()

    log.Println("🔧 Opening database connection...")
    database.Connect()

    log.Println("✅ Initialization complete!")
}

func main() {
    go func() {
        database.RunMigrations()
    }()

    defer database.Close()

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    router := route.SetupRouter()

    log.Println("🚀 Starting server on port:", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("❌ Failed to start server: %v", err)
    }
}