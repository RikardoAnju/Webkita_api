package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/database"
	"BackendFramework/internal/service"
	"github.com/gin-gonic/gin"
)

func InitProjectRoutes(r *gin.RouterGroup) {
	project := r.Group("/project")
	
	// Inisialisasi service dan controller
	projectService := service.NewProjectService(database.DbWebkita)
	projectCtrl := controller.NewProjectController(projectService)

	{
		// Routes tanpa trailing slash
		project.POST("", projectCtrl.CreateProject)
		project.GET("", projectCtrl.GetAllProjects)
		project.GET("/:projectId", projectCtrl.GetProjectByID)
		project.GET("/user/:userId", projectCtrl.GetProjectsByUserID)
		project.PATCH("/:projectId/status", projectCtrl.UpdateProjectStatus)
		project.DELETE("/:projectId", projectCtrl.DeleteProject)
	}
}