package config

import (
    "log"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
    dsn := os.Getenv("DATABASE_URL")
	 log.Println("🔍 DSN:", dsn)
    if dsn == "" {
        log.Fatal("❌ DATABASE_URL is not set")
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        log.Fatalf("❌ Failed to connect to database: %v", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalf("❌ Failed to get underlying sql.DB: %v", err)
    }

    if err := sqlDB.Ping(); err != nil {
        log.Fatalf("❌ Failed to ping database: %v", err)
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)

    log.Println("✅ Connected to Supabase PostgreSQL")
    DB = db
}

func RunMigrations() {
    log.Println("🔄 Starting auto-migration...")
    err := DB.AutoMigrate(
        // &model.User{},
        // &model.Project{},
    )
    if err != nil {
        log.Fatalf("❌ Migration failed: %v", err)
    }
    log.Println("✅ Migration completed")
}

func Close() {
    sqlDB, err := DB.DB()
    if err != nil {
        log.Printf("❌ Failed to get sql.DB: %v", err)
        return
    }
    if err := sqlDB.Close(); err != nil {
        log.Printf("❌ Failed to close database: %v", err)
        return
    }
    log.Println("✅ Database connection closed")
}