package v1

import (
    "BackendFramework/internal/controller"
    "BackendFramework/internal/database"
    "BackendFramework/internal/middleware"
    "BackendFramework/internal/model"
    "BackendFramework/internal/service"
    "github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.RouterGroup) {
    // Public Routes (Authentication)
    auth := r.Group("/auth")
    {
        auth.POST("/login", controller.Login)
        auth.POST("/login-email", controller.NewAuthController().LoginWithEmail)
        auth.POST("/login-username", controller.NewAuthController().LoginWithUsername)
        auth.POST("/register", controller.NewAuthController().Register)
        auth.GET("/logout/:usrId", controller.Logout)
        auth.POST("/refresh-access", middleware.LogUserActivity(), controller.RefreshAccessToken)
    }

    // User Group
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

    
    project := r.Group("/project")
   
    {
        // Inisialisasi service dan controller
        projectService := service.NewProjectService(database.DbWebkita)
        projectCtrl := controller.NewProjectController(projectService)
        
        // Routes tanpa trailing slash
        project.POST("", projectCtrl.CreateProject)
        project.GET("", projectCtrl.GetAllProjects)
        project.GET("/:projectId", projectCtrl.GetProjectByID)
        project.GET("/user/:userId", projectCtrl.GetProjectsByUserID)
        project.PATCH("/:projectId/status", projectCtrl.UpdateProjectStatus)
        project.DELETE("/:projectId", projectCtrl.DeleteProject)
    }

    // Misc Group
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