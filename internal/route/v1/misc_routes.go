package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"github.com/gin-gonic/gin"
)

func InitMiscRoutes(r *gin.RouterGroup) {
	misc := r.Group("/misc")
	misc.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
	{
		fileInput := &model.FileInput{}
		misc.POST("/upload-data-s3-local", middleware.InputValidator(fileInput), controller.UploadFile)
		misc.GET("/generate-pdf", controller.TryGeneratePdf)
		misc.GET("/send-mail", controller.SendMail)
		misc.GET("/generate-excel", controller.GenerateExcel)
		misc.POST("/read-excel", controller.ReadExcel)
		misc.GET("/test-ping", controller.PingMongo)
	}
}