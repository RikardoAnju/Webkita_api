package database

import (
	"log"

	"BackendFramework/internal/config"
	"BackendFramework/internal/model"
	"gorm.io/gorm"
)

// DbWebkita adalah alias ke koneksi utama Supabase
var DbWebkita *gorm.DB

func Connect() {
	config.Connect()
	DbWebkita = config.DB
	log.Println("✅ DbWebkita ready")
}

func RunMigrations() {
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