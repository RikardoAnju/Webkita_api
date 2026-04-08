// internal/controller/project_controller.go
package controller

import (
    "BackendFramework/internal/model"
    "BackendFramework/internal/service"
    "github.com/gin-gonic/gin"
    "log"
    "net/http"
    "strconv"
)

type ProjectController struct {
    service *service.ProjectService
}

func NewProjectController(service *service.ProjectService) *ProjectController {
    return &ProjectController{service: service}
}

func (ctrl *ProjectController) CreateProject(c *gin.Context) {
    log.Println("=== START CREATE PROJECT ===")
    
    // Parse multipart form (max 32 MB)
    if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
        log.Printf("❌ Failed to parse multipart form: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Failed to parse form",
            "message": err.Error(),
        })
        return
    }

    // Log semua form values untuk debugging
    log.Println("📋 Form values received:")
    for key, values := range c.Request.PostForm {
        log.Printf("  %s: %v", key, values)
    }

    // Parse userId
    userIDStr := c.PostForm("userId")
    if userIDStr == "" {
        log.Println("❌ userId is missing")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "userId is required",
        })
        return
    }

    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        log.Printf("❌ Invalid userId format: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid userId",
            "message": err.Error(),
        })
        return
    }

    // Manual binding untuk multipart form
    input := model.ProjectCreateInput{
        UserID:          userID,
        PlanTitle:       c.PostForm("planTitle"),
        ProjectTitle:    c.PostForm("projectTitle"),
        Category:        c.PostForm("category"),
        Description:     c.PostForm("description"),
        Skills:          c.PostForm("skills"),
        ContactName:     c.PostForm("contactName"),
        ContactPhone:    c.PostForm("contactPhone"),
        AdditionalNotes: c.PostForm("additionalNotes"),
    }

    // Log parsed input
    log.Printf("📥 Parsed input: UserID=%d, PlanTitle=%s, ProjectTitle=%s, Category=%s", 
        input.UserID, input.PlanTitle, input.ProjectTitle, input.Category)
    log.Printf("📝 Description length: %d chars", len(input.Description))
    log.Printf("🏷️  Skills: %s", input.Skills)
    log.Printf("📞 Contact: %s (%s)", input.ContactName, input.ContactPhone)

    // Handle file attachments
    form, err := c.MultipartForm()
    if err == nil && form.File["attachments"] != nil {
        files := form.File["attachments"]
        log.Printf("📎 Received %d attachment(s)", len(files))
        
        for i, file := range files {
            log.Printf("  File %d: %s (%.2f MB)", i+1, file.Filename, float64(file.Size)/(1024*1024))
        }
        
        input.Attachments = files
    } else {
        log.Println("📎 No attachments received")
    }

    // Validasi manual
    if input.ProjectTitle == "" {
        log.Println("❌ Validation failed: projectTitle is empty")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "projectTitle is required",
        })
        return
    }

    if input.Category == "" {
        log.Println("❌ Validation failed: category is empty")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "category is required",
        })
        return
    }

    if input.Description == "" || len(input.Description) < 100 {
        log.Printf("❌ Validation failed: description too short (%d chars)", len(input.Description))
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "description must be at least 100 characters",
        })
        return
    }

    if input.Skills == "" {
        log.Println("❌ Validation failed: skills is empty")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "skills is required",
        })
        return
    }

    if input.ContactName == "" {
        log.Println("❌ Validation failed: contactName is empty")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "contactName is required",
        })
        return
    }

    if input.ContactPhone == "" {
        log.Println("❌ Validation failed: contactPhone is empty")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "contactPhone is required",
        })
        return
    }

    log.Println("✅ Validation passed, calling service...")

    // Create project
    project, err := ctrl.service.CreateProject(&input)
    if err != nil {
        log.Printf("❌ Service error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Failed to create project",
            "message": err.Error(),
        })
        return
    }

    log.Printf("✅ Project created successfully - ID: %d", project.ID)
    log.Println("=== END CREATE PROJECT ===")

    c.JSON(http.StatusOK, gin.H{
        "message":   "Project created successfully",
        "projectId": project.ID,
        "data":      project,
    })
}

func (ctrl *ProjectController) GetProjectByID(c *gin.Context) {
    log.Printf("📖 Getting project by ID: %s", c.Param("projectId"))
    
    id, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
    if err != nil {
        log.Printf("❌ Invalid project ID: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    project, err := ctrl.service.GetProjectByID(uint(id))
    if err != nil {
        log.Printf("❌ Project not found: %v", err)
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    log.Printf("✅ Project found: %d", project.ID)
    c.JSON(http.StatusOK, gin.H{"data": project})
}

func (ctrl *ProjectController) GetProjectsByUserID(c *gin.Context) {
    userIDStr := c.Param("userId")
    log.Printf("📖 Getting projects for user: %s", userIDStr)
    
    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        log.Printf("❌ Invalid user ID: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    projects, err := ctrl.service.GetProjectsByUserID(userID)
    if err != nil {
        log.Printf("❌ Error fetching projects: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    log.Printf("✅ Found %d project(s) for user %d", len(projects), userID)
    c.JSON(http.StatusOK, gin.H{
        "data":  projects,
        "count": len(projects),
    })
}

func (ctrl *ProjectController) GetAllProjects(c *gin.Context) {
    log.Println("📖 Getting all projects")
    
    projects, err := ctrl.service.GetAllProjects()
    if err != nil {
        log.Printf("❌ Error fetching projects: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    log.Printf("✅ Found %d project(s)", len(projects))
    c.JSON(http.StatusOK, gin.H{
        "data":  projects,
        "count": len(projects),
    })
}

func (ctrl *ProjectController) UpdateProjectStatus(c *gin.Context) {
    projectIDStr := c.Param("projectId")
    log.Printf("🔄 Updating project status: %s", projectIDStr)
    
    id, err := strconv.ParseUint(projectIDStr, 10, 32)
    if err != nil {
        log.Printf("❌ Invalid project ID: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    var input struct {
        Status string `json:"status" binding:"required"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        log.Printf("❌ Invalid request body: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    log.Printf("🔄 Changing project %d status to: %s", id, input.Status)

    if err := ctrl.service.UpdateProjectStatus(uint(id), input.Status); err != nil {
        log.Printf("❌ Failed to update status: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    log.Printf("✅ Project %d status updated successfully", id)
    c.JSON(http.StatusOK, gin.H{
        "message": "Project status updated successfully",
        "status":  input.Status,
    })
}

func (ctrl *ProjectController) DeleteProject(c *gin.Context) {
    projectIDStr := c.Param("projectId")
    log.Printf("🗑️  Deleting project: %s", projectIDStr)
    
    id, err := strconv.ParseUint(projectIDStr, 10, 32)
    if err != nil {
        log.Printf("❌ Invalid project ID: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    if err := ctrl.service.DeleteProject(uint(id)); err != nil {
        log.Printf("❌ Failed to delete project: %v", err)
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    log.Printf("✅ Project %d deleted successfully", id)
    c.JSON(http.StatusOK, gin.H{
        "message": "Project deleted successfully",
    })
}