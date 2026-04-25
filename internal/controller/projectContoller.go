package controller

import (
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ─── POST /v1/project ──────────────────────────────────────────────────────────

func CreateProject(c *gin.Context) {
	
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Unauthorized"})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Invalid user ID in token"})
		return
	}

	userIDParsed, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Invalid user ID format"})
		return
	}

	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "error": "Failed to parse form data: " + err.Error()})
		return
	}

	submission := &model.ProjectSubmission{
		UserID:          uint(userIDParsed),
		PlanTitle:       c.PostForm("planTitle"),
		ProjectTitle:    c.PostForm("projectTitle"),
		Category:        c.PostForm("category"),
		Description:     c.PostForm("description"),
		Skills:          c.PostForm("skills"),
		ContactName:     c.PostForm("contactName"),
		ContactPhone:    c.PostForm("contactPhone"),
		AdditionalNotes: c.PostForm("additionalNotes"),
	}

	var files []*multipart.FileHeader
	if form := c.Request.MultipartForm; form != nil && form.File["attachments"] != nil {
		files = form.File["attachments"]
	}

	response, err := service.CreateProject(submission, files)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnprocessableEntity, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "message": "Project submitted successfully", "data": response})
}

// ─── GET /v1/project ───────────────────────────────────────────────────────────

func GetAllProjects(c *gin.Context) {
	projects, err := service.GetAllProjects()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Projects retrieved successfully",
		"data":    projects,
		"total":   len(projects),
	})
}

// ─── GET /v1/project/my ────────────────────────────────────────────────────────

func GetMyProjects(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Unauthorized"})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Invalid user ID in token"})
		return
	}

	userIDParsed, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Invalid user ID format"})
		return
	}

	projects, err := service.GetProjectsByUserID(uint(userIDParsed))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Projects retrieved successfully",
		"data":    projects,
		"total":   len(projects),
	})
}

// ─── GET /v1/project/:id ───────────────────────────────────────────────────────

func GetProjectByID(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	project, err := service.GetProjectByID(projectID)
	if err != nil {
		code := http.StatusInternalServerError
		if containsString(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "error": err.Error()})
		return
	}

	attachments, _ := service.GetAttachmentsByProjectID(projectID)

	c.JSON(http.StatusOK, gin.H{
		"code":        http.StatusOK,
		"message":     "Project retrieved successfully",
		"project":     project,
		"attachments": attachments,
	})
}

// ─── GET /v1/project/user/:userId ─────────────────────────────────────────────

func GetProjectsByUser(c *gin.Context) {
	userID, err := parseUintParam(c, "userId")
	if err != nil {
		return
	}

	projects, err := service.GetProjectsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Projects retrieved successfully",
		"data":    projects,
		"total":   len(projects),
	})
}

// ─── PATCH /v1/project/:id/status ─────────────────────────────────────────────

func UpdateProjectStatus(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "error": "Field 'status' is required"})
		return
	}

	if err := service.UpdateProjectStatus(projectID, model.ProjectStatus(body.Status)); err != nil {
		code := http.StatusInternalServerError
		if containsString(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		if err == model.ErrInvalidProjectStatus {
			code = http.StatusBadRequest
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Project status updated successfully",
		"status":  body.Status,
	})
}

// ─── DELETE /v1/project/:id ────────────────────────────────────────────────────

func DeleteProject(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	if err := service.DeleteProject(projectID); err != nil {
		code := http.StatusInternalServerError
		if containsString(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "Project deleted successfully"})
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

func parseUintParam(c *gin.Context, param string) (uint, error) {
	raw := c.Param(param)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "error": param + " is not valid"})
		return 0, err
	}
	return uint(id), nil
}

func containsString(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
