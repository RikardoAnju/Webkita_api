package v1

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes adalah entry point utama untuk semua API v1
func SetupRoutes(r *gin.Engine) {
	// --- Middleware CORS Global ---
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle Preflight OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// --- Grouping API v1 ---
	apiV1 := r.Group("/api/v1")
	{
		InitAuthRoutes(apiV1)    // /api/v1/auth/...
		InitUserRoutes(apiV1)    // /api/v1/user/...
		InitProjectRoutes(apiV1) // /api/v1/project/...
		InitMiscRoutes(apiV1)    // /api/v1/misc/...
	}
}