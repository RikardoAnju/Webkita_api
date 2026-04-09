package service

import (
    "log"
    "BackendFramework/internal/database"
)

func TestPing() {
    sqlDB, err := database.DbWebkita.DB()
    if err != nil {
        log.Println("❌ Ping Failed - cannot get DB instance:", err)
        return
    }

    if err := sqlDB.Ping(); err != nil {
        log.Println("❌ Ping Failed:", err)
        return
    }

    log.Println("✅ DB Ping successful")
}