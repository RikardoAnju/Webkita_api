package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ─── Project CRUD ──────────────────────────────────────────────────────────────

// GetAllProjects mengambil semua project
func GetAllProjects() ([]model.Project, error) {
	var projects []model.Project

	err := database.DbWebkita.
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&projects).Error

	if err != nil {
		middleware.LogError(err, "Query Error: GetAllProjects")
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	return projects, nil
}

// GetProjectByID mengambil satu project berdasarkan ID
func GetProjectByID(projectID uint) (*model.Project, error) {
	if projectID == 0 {
		return nil, errors.New("project ID cannot be zero")
	}

	var project model.Project

	err := database.DbWebkita.
		Where("id = ? AND deleted_at IS NULL", projectID).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project not found with ID: %d", projectID)
		}
		middleware.LogError(err, fmt.Sprintf("Data Scan Error: GetProjectByID for ID %d", projectID))
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetProjectsByUserID mengambil semua project milik user tertentu
func GetProjectsByUserID(userID uint) ([]model.Project, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	var projects []model.Project

	err := database.DbWebkita.
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&projects).Error

	if err != nil {
		middleware.LogError(err, fmt.Sprintf("Query Error: GetProjectsByUserID for userID %d", userID))
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	return projects, nil
}

// GetAttachmentsByProjectID mengambil semua attachment milik suatu project
func GetAttachmentsByProjectID(projectID uint) ([]model.ProjectAttachment, error) {
	if projectID == 0 {
		return nil, errors.New("project ID cannot be zero")
	}

	var attachments []model.ProjectAttachment

	err := database.DbWebkita.
		Where("project_id = ?", projectID).
		Find(&attachments).Error

	if err != nil {
		middleware.LogError(err, fmt.Sprintf("Query Error: GetAttachmentsByProjectID for projectID %d", projectID))
		return nil, fmt.Errorf("failed to query attachments: %w", err)
	}

	return attachments, nil
}

// CreateProject menyimpan submission proyek beserta attachment-nya ke database
func CreateProject(submission *model.ProjectSubmission, files []*multipart.FileHeader) (*model.ProjectResponse, error) {
	if submission == nil {
		return nil, errors.New("submission data cannot be nil")
	}

	// Validasi field wajib
	if submission.UserID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}
	if submission.ProjectTitle == "" {
		return nil, errors.New("project title cannot be empty")
	}
	if submission.Category == "" {
		return nil, errors.New("category cannot be empty")
	}
	if len(submission.Description) < 100 {
		return nil, errors.New("description must be at least 100 characters")
	}
	if submission.Skills == "" {
		return nil, errors.New("skills cannot be empty")
	}
	if submission.ContactName == "" {
		return nil, errors.New("contact name cannot be empty")
	}
	if submission.ContactPhone == "" {
		return nil, errors.New("contact phone cannot be empty")
	}

	// Validasi kategori
	if !model.IsValidCategory(submission.Category) {
		return nil, model.ErrInvalidCategory
	}

	// Buat entitas Project
	project := &model.Project{
		UserID:          submission.UserID,
		PlanTitle:       submission.PlanTitle,
		ProjectTitle:    submission.ProjectTitle,
		Category:        submission.Category,
		Description:     submission.Description,
		Skills:          submission.Skills,
		ContactName:     submission.ContactName,
		ContactPhone:    submission.ContactPhone,
		AdditionalNotes: submission.AdditionalNotes,
		Status:          string(model.StatusPending),
	}

	// Jalankan dalam transaksi agar project + attachment atomic
	var savedAttachments []model.ProjectAttachment

	err := database.DbWebkita.Transaction(func(tx *gorm.DB) error {
		// Simpan project
		if err := tx.Create(project).Error; err != nil {
			middleware.LogError(err, "Insert Data Failed: CreateProject")
			return fmt.Errorf("failed to insert project: %w", err)
		}

		// Proses dan simpan attachment (opsional)
		for _, fileHeader := range files {
			attachment, err := saveAttachment(project.ID, fileHeader)
			if err != nil {
				return fmt.Errorf("failed to save file %s: %w", fileHeader.Filename, err)
			}

			if err := tx.Create(attachment).Error; err != nil {
				middleware.LogError(err, "Insert Data Failed: CreateProjectAttachment")
				return fmt.Errorf("failed to insert attachment record: %w", err)
			}

			savedAttachments = append(savedAttachments, *attachment)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &model.ProjectResponse{
		ID:           project.ID,
		ProjectTitle: project.ProjectTitle,
		Category:     project.Category,
		Status:       project.Status,
		Attachments:  savedAttachments,
		CreatedAt:    project.CreatedAt,
		Message:      "Proyek berhasil diajukan. Tim kami akan menghubungi Anda segera.",
	}, nil
}

// UpdateProjectStatus mengubah status project
func UpdateProjectStatus(projectID uint, status model.ProjectStatus) error {
	if projectID == 0 {
		return errors.New("project ID cannot be zero")
	}

	// Validasi nilai status
	if !model.IsValidProjectStatus(status) {
		return model.ErrInvalidProjectStatus
	}

	result := database.DbWebkita.
		Table("projects").
		Where("id = ? AND deleted_at IS NULL", projectID).
		Updates(map[string]interface{}{
			"status":     string(status),
			"updated_at": gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		middleware.LogError(result.Error, "Update Data Failed: UpdateProjectStatus")
		return fmt.Errorf("failed to update project status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("project not found or no changes made")
	}

	return nil
}

// DeleteProject soft delete project
func DeleteProject(projectID uint) error {
	if projectID == 0 {
		return errors.New("project ID cannot be zero")
	}

	result := database.DbWebkita.
		Table("projects").
		Where("id = ? AND deleted_at IS NULL", projectID).
		Update("deleted_at", gorm.Expr("NOW()"))

	if result.Error != nil {
		middleware.LogError(result.Error, "Delete Data Failed: DeleteProject")
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("project not found")
	}

	return nil
}

// CheckProjectExists memeriksa apakah project dengan ID sudah ada
func CheckProjectExists(projectID uint) (bool, error) {
	var count int64

	err := database.DbWebkita.
		Table("projects").
		Where("id = ? AND deleted_at IS NULL", projectID).
		Count(&count).Error

	if err != nil {
		middleware.LogError(err, "Check Project Exists Failed")
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return count > 0, nil
}

// ─── Private helpers ───────────────────────────────────────────────────────────

// saveAttachment menyimpan file ke disk dan mengembalikan model attachment
func saveAttachment(projectID uint, fileHeader *multipart.FileHeader) (*model.ProjectAttachment, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Buat direktori upload
	uploadDir := fmt.Sprintf("uploads/projects/%d", projectID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Buat nama file unik
	ext := filepath.Ext(fileHeader.Filename)
	baseName := strings.TrimSuffix(filepath.Base(fileHeader.Filename), ext)
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)
	storagePath := filepath.Join(uploadDir, uniqueName)

	// Tulis file ke disk
	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return &model.ProjectAttachment{
		ProjectID:   projectID,
		FileName:    fileHeader.Filename,
		FileSize:    fileHeader.Size,
		FileType:    fileHeader.Header.Get("Content-Type"),
		StoragePath: storagePath,
	}, nil
}