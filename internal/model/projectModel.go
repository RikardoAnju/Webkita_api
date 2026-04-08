package model

import (
    "log"
    "mime/multipart"
    "gorm.io/gorm"
)

type Project struct {
    gorm.Model
    UserID          int    `gorm:"not null;index" json:"user_id"`
    ProjectTitle    string `gorm:"type:varchar(255);not null" json:"project_title"`
    Category        string `gorm:"type:varchar(100);not null" json:"category"`
    Description     string `gorm:"type:text;not null" json:"description"`
    Skills          string `gorm:"type:text" json:"skills"`
    ContactName     string `gorm:"type:varchar(255);not null" json:"contact_name"`
    ContactPhone    string `gorm:"type:varchar(50);not null" json:"contact_phone"`
    AdditionalNotes string `gorm:"type:text" json:"additional_notes"`
    PlanTitle       string `gorm:"type:varchar(255)" json:"plan_title"`
    PlanPriceRange  string `gorm:"type:varchar(100)" json:"plan_price_range"`
    Status          string `gorm:"type:varchar(50);default:'pending'" json:"status"`
}

func (Project) TableName() string {
    return "projects"
}

// GORM Hooks for debugging
func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
    log.Println(" HOOK: BeforeCreate called")
    log.Printf("   UserID: %d", p.UserID)
    log.Printf("   ProjectTitle: %s", p.ProjectTitle)
    log.Printf("   Category: %s", p.Category)
    log.Printf("   Status: %s", p.Status)
    
    
    var userExists bool
    err = tx.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL)", p.UserID).Scan(&userExists).Error
    if err != nil {
        log.Printf(" HOOK: Error checking user existence: %v", err)
        return err
    }
    
    if !userExists {
        log.Printf(" HOOK: User %d does not exist!", p.UserID)
        return gorm.ErrRecordNotFound
    }
    
    log.Printf(" HOOK: User %d exists", p.UserID)
    return nil
}

func (p *Project) AfterCreate(tx *gorm.DB) (err error) {
    log.Println(" HOOK: AfterCreate called")
    log.Printf("   Project ID after insert: %d", p.ID)
    log.Printf("   Created At: %v", p.CreatedAt)
    

    var count int64
    tx.Model(&Project{}).Where("id = ?", p.ID).Count(&count)
    log.Printf("   Records found with ID %d: %d", p.ID, count)
    
    if count == 0 {
        log.Println(" HOOK: CRITICAL - Record not found after insert!")
    } else {
        log.Println(" HOOK: Record verified in database")
    }
    
    return nil
}

func (p *Project) BeforeSave(tx *gorm.DB) (err error) {
    log.Println(" HOOK: BeforeSave called")
    return nil
}

func (p *Project) AfterSave(tx *gorm.DB) (err error) {
    log.Println(" HOOK: AfterSave called")
    return nil
}

// Input struct for multipart form
type ProjectCreateInput struct {
    UserID          int                     `form:"userId" binding:"required"`
    ProjectTitle    string                  `form:"projectTitle" binding:"required"`
    Category        string                  `form:"category" binding:"required"`
    Description     string                  `form:"description" binding:"required,min=100"`
    Skills          string                  `form:"skills" binding:"required"`
    ContactName     string                  `form:"contactName" binding:"required"`
    ContactPhone    string                  `form:"contactPhone" binding:"required"`
    AdditionalNotes string                  `form:"additionalNotes"`
    PlanTitle       string                  `form:"planTitle"`
    Attachments     []*multipart.FileHeader `form:"attachments"`
}