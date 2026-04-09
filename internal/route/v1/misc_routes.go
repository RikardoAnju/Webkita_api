package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitMiscRoutes(r *gin.RouterGroup, db *gorm.DB) {
	misc := r.Group("/misc")
	misc.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
	{
	
		misc.GET("/generate-pdf", controller.TryGeneratePdf)
		misc.GET("/send-mail", controller.SendMail)
		misc.GET("/generate-excel", controller.GenerateExcel)
		misc.POST("/read-excel", controller.ReadExcel)
		misc.GET("/test-ping", controller.PingMongo)
	}
}