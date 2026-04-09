package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitProjectRoutes(r *gin.RouterGroup, db *gorm.DB) {
	project := r.Group("/project")

	projectService := service.NewProjectService(db) // pakai db dari parameter, bukan hardcoded
	projectCtrl := controller.NewProjectController(projectService)
	{
		project.POST("", projectCtrl.CreateProject)
		project.GET("", projectCtrl.GetAllProjects)
		project.GET("/:projectId", projectCtrl.GetProjectByID)
		project.GET("/user/:userId", projectCtrl.GetProjectsByUserID)
		project.PATCH("/:projectId/status", projectCtrl.UpdateProjectStatus)
		project.DELETE("/:projectId", projectCtrl.DeleteProject)
	}
}