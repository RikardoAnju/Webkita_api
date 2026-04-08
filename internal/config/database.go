package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("❌ DATABASE_URL not found")
	}

	log.Println("🔧 Opening database connection...")
	log.Println("🔍 DSN:", dsn)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false, // ✅ FIX: disable prepared statement (WAJIB untuk Supabase)
	})

	if err != nil {
		log.Fatalf("❌ Failed to connect database: %v", err)
	}

	log.Println("✅ Connected to Supabase PostgreSQL")

	// ✅ tambahan safety (optional tapi bagus)
	DB = DB.Session(&gorm.Session{
		PrepareStmt: false,
	})
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Println("❌ Failed to get sqlDB:", err)
		return
	}
	sqlDB.Close()
}

// helper env
func GetEnv(key string) string {
	return os.Getenv(key)
}