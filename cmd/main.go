package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin" // Tambahkan Gin di sini
	"BackendFramework/internal/config"
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/route/v1" // Package ini dipanggil sebagai 'v1'
)

func init() {
	// Load .env dulu via InitEnvironment
	config.InitEnvronment()

	log.Println("🔧 Initializing middleware...")
	middleware.InitLogger()
	middleware.InitValidator()

	log.Println("🔧 Opening database connection...")
	database.Connect()

	log.Println("✅ Initialization complete!")
}

func main() {
	// 1. JALANKAN SECARA SINKRON: Tunggu migrasi selesai baru lanjut ke kode bawahnya
	log.Println("🛠️ Running database migrations...")
	database.RunMigrations()

	// Tutup koneksi saat aplikasi mati
	defer database.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. INISIALISASI ROUTER
	router := gin.Default()

	// 3. DAFTARKAN ROUTES V1 YANG SUDAH KITA PISAH TADI
	// Biasanya kita buatkan prefix/group seperti /api/v1
	apiV1 := router.Group("/api/v1")
	v1.InitRoutes(apiV1) 

	log.Println("🚀 Starting server on port:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}