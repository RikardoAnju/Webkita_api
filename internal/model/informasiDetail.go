package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ─── Errors ────────────────────────────────────────────────────────────────────

var (
	ErrInvalidCategory      = errors.New("invalid category")
	ErrInvalidProjectStatus = errors.New("invalid project status. use: pending | in_review | approved | rejected")
)

// ─── Enums / Konstanta ─────────────────────────────────────────────────────────

type ProjectStatus string

const (
	StatusPending   ProjectStatus = "pending"
	StatusInReview  ProjectStatus = "in_review"
	StatusApproved  ProjectStatus = "approved"
	StatusRejected  ProjectStatus = "rejected"
)

type CategoryType string

const (
	CategoryEcommerce      CategoryType = "E-commerce"
	CategoryLandingPage    CategoryType = "Landing Page"
	CategoryCompanyProfile CategoryType = "Company Profile"
	CategoryMobileApp      CategoryType = "Mobile Application"
	CategoryDashboard      CategoryType = "Dashboard"
	CategoryWebApp         CategoryType = "Web Application"
	CategoryBlogCMS        CategoryType = "Blog/CMS"
	CategoryPortfolio      CategoryType = "Portfolio"
	CategoryOther          CategoryType = "Lainnya"
)

var validCategories = map[string]bool{
	string(CategoryEcommerce):      true,
	string(CategoryLandingPage):    true,
	string(CategoryCompanyProfile): true,
	string(CategoryMobileApp):      true,
	string(CategoryDashboard):      true,
	string(CategoryWebApp):         true,
	string(CategoryBlogCMS):        true,
	string(CategoryPortfolio):      true,
	string(CategoryOther):          true,
}

var validStatuses = map[ProjectStatus]bool{
	StatusPending:  true,
	StatusInReview: true,
	StatusApproved: true,
	StatusRejected: true,
}

func IsValidCategory(cat string) bool {
	return validCategories[cat]
}

func IsValidProjectStatus(status ProjectStatus) bool {
	return validStatuses[status]
}

// ─── Database Models ───────────────────────────────────────────────────────────

// Project adalah model utama yang disimpan ke tabel "projects"
type Project struct {
	ID              uint           `json:"id"               gorm:"primaryKey;autoIncrement"`
	UserID          uint           `json:"userId"           gorm:"not null;index"`
	PlanTitle       string         `json:"planTitle"        gorm:"type:varchar(100)"`
	ProjectTitle    string         `json:"projectTitle"     gorm:"type:varchar(200);not null"`
	Category        string         `json:"category"         gorm:"type:varchar(100);not null"`
	Description     string         `json:"description"      gorm:"type:text;not null"`
	Skills          string         `json:"skills"           gorm:"type:text;not null"` 
	ContactName     string         `json:"contactName"      gorm:"type:varchar(100);not null"`
	ContactPhone    string         `json:"contactPhone"     gorm:"type:varchar(20);not null"`
	AdditionalNotes string         `json:"additionalNotes"  gorm:"type:text"`
	Status          string         `json:"status"           gorm:"type:varchar(20);default:'pending'"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `json:"-"                gorm:"index"`
}

// ProjectAttachment adalah model untuk file lampiran
type ProjectAttachment struct {
	ID          uint      `json:"id"          gorm:"primaryKey;autoIncrement"`
	ProjectID   uint      `json:"projectId"   gorm:"not null;index"`
	FileName    string    `json:"fileName"    gorm:"type:varchar(255);not null"`
	FileSize    int64     `json:"fileSize"    gorm:"not null"`
	FileType    string    `json:"fileType"    gorm:"type:varchar(100)"`
	StoragePath string    `json:"storagePath" gorm:"type:varchar(500);not null"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ─── Input / DTO ───────────────────────────────────────────────────────────────

// ProjectSubmission adalah input dari form multipart/form-data
type ProjectSubmission struct {
	UserID          uint   // diambil dari JWT claims, bukan dari form
	PlanTitle       string `form:"planTitle"`
	ProjectTitle    string `form:"projectTitle"`
	Category        string `form:"category"`
	Description     string `form:"description"`
	Skills          string `form:"skills"` // comma-separated: "React, Node.js, MongoDB"
	ContactName     string `form:"contactName"`
	ContactPhone    string `form:"contactPhone"`
	AdditionalNotes string `form:"additionalNotes"`
}

// ProjectResponse adalah response yang dikembalikan setelah submission berhasil
type ProjectResponse struct {
	ID           uint                `json:"id"`
	ProjectTitle string              `json:"projectTitle"`
	Category     string              `json:"category"`
	Status       string              `json:"status"`
	Attachments  []ProjectAttachment `json:"attachments,omitempty"`
	CreatedAt    time.Time           `json:"createdAt"`
	Message      string              `json:"message"`
}