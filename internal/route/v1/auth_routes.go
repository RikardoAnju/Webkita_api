package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitAuthRoutes(r *gin.RouterGroup, db *gorm.DB) {
	auth := r.Group("/auth")

	authCtrl := controller.NewAuthController(db)
	{
		auth.POST("/login-email", authCtrl.LoginWithEmail)
		auth.POST("/login-username", authCtrl.LoginWithUsername)
		auth.POST("/register", authCtrl.Register)
		auth.GET("/verify-email", authCtrl.VerifyEmail)
		auth.POST("/resend-verification", authCtrl.ResendVerification)
	}

	authProtected := auth.Group("")
	authProtected.Use(middleware.LogUserActivity())
	{
		authProtected.GET("/profile", authCtrl.GetProfile)
	}
}