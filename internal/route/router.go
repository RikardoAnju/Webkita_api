package route

import (
    "os"
    "github.com/gin-gonic/gin"
    "BackendFramework/internal/controller"
    "BackendFramework/internal/route/v1"
)

func SetupRouter() *gin.Engine {
   
    if os.Getenv("ENVIRONMENT") == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    r := gin.Default()
   
    // ========== CUSTOM CORS MIDDLEWARE ==========
    r.Use(func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        
        // Allowed origins
        allowedOrigins := map[string]bool{
            "http://localhost:3000": true,
            "http://localhost:5173": true,
            "http://localhost:5174": true,
        }
        
        // Set CORS headers jika origin diizinkan
        if allowedOrigins[origin] {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
            c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
            c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
            c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
            c.Writer.Header().Set("Access-Control-Max-Age", "43200")
        }
        
        // Handle preflight
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    })
    // ============================================
   
    // Route /api/auth
    api := r.Group("/api")
    {
        auth := api.Group("/auth")
        authController := controller.NewAuthController()
        auth.POST("/register", authController.Register)
        auth.POST("/login", authController.LoginWithEmail)
        auth.POST("/login-username", authController.LoginWithUsername)
    }
   
    // V1 routes
    v1Routes := r.Group("/v1")
    {
        v1.InitRoutes(v1Routes)
    }
   
    return r
}