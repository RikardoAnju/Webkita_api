package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/middleware"

	"github.com/gin-gonic/gin"
)

func ProjectRoutes(r *gin.RouterGroup) {
	project := r.Group("/project")
	project.Use(middleware.JWTAuthMiddleware())
	{
		// User routes (semua user login)
		project.POST("", controller.CreateProject)      // POST   /v1/project      - ajukan project baru
		project.GET("/my", controller.GetMyProjects)    // GET    /v1/project/my   - project milik saya
		project.GET("/:id", controller.GetProjectByID)  // GET    /v1/project/:id  - detail project

		// Admin only routes
		admin := project.Group("")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("", controller.GetAllProjects)                   // GET    /v1/project
			admin.GET("/user/:userId", controller.GetProjectsByUser)   // GET    /v1/project/user/:userId
			admin.PATCH("/:id/status", controller.UpdateProjectStatus) // PATCH  /v1/project/:id/status
			admin.DELETE("/:id", controller.DeleteProject)             // DELETE /v1/project/:id
		}
	}
}