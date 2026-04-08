package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"github.com/gin-gonic/gin"
)

func InitUserRoutes(r *gin.RouterGroup) {
	user := r.Group("/user")
	user.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
	{
		user.GET("/", controller.GetUser)
		user.GET("/:usrId", controller.GetUser)
		user.DELETE("/:usrId", controller.DeleteUser)
		
		userInput := &model.UserInput{}
		user.PUT("/", middleware.InputValidator(userInput), controller.InsertUser)
		user.PATCH("/", middleware.InputValidator(userInput), controller.UpdateUser)
		user.GET("/profile", controller.GetUserProfile)
	}
}