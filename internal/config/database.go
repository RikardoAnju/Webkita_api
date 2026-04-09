package config

import (
	"log"
	"os"
	"time"

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
		PrepareStmt: false, // ✅ WAJIB untuk Supabase
	})

	if err != nil {
		log.Fatalf("❌ Failed to connect database: %v", err)
	}

	log.Println("✅ Connected to Supabase PostgreSQL")

	
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("❌ Failed to get sqlDB:", err)
	}

	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// ✅ safety extra
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

func GetEnv(key string) string {
	return os.Getenv(key)
}