package database

import (
	"log"
	"os"

	"BackendFramework/internal/config"
	"BackendFramework/internal/model"
	"gorm.io/gorm"
)

var DbWebkita *gorm.DB

func Connect() {
	// ✅ FIX logging (biar gak error read-only filesystem)
	log.SetOutput(os.Stdout)

	config.Connect()
	DbWebkita = config.DB

	log.Println("✅ DbWebkita ready")
}

func RunMigrations() {
	// ✅ FIX: biar gak jalan tiap startup
	if config.GetEnv("RUN_MIGRATION") != "true" {
		log.Println("⏭️ Skipping migration")
		return
	}

	log.Println("🔄 Starting auto-migration...")

	err := DbWebkita.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.UserToken{},
	)

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migration completed")
}

func Close() {
	config.Close()
}
