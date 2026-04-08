package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/middleware"
	"github.com/gin-gonic/gin"
)

func InitAuthRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	
	// Praktik terbaik: Inisialisasi controller satu kali saja di luar scope rute
	// daripada memanggil NewAuthController() berkali-kali.
	authCtrl := controller.NewAuthController()

	{
		auth.POST("/login", controller.Login)
		auth.POST("/login-email", authCtrl.LoginWithEmail)
		auth.POST("/login-username", authCtrl.LoginWithUsername)
		auth.POST("/register", authCtrl.Register)
		auth.GET("/logout/:usrId", controller.Logout)
		auth.POST("/refresh-access", middleware.LogUserActivity(), controller.RefreshAccessToken)
	}
}