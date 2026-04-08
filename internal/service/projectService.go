
package service

import (
    "BackendFramework/internal/model"
    "errors"
    "gorm.io/gorm"
)

type ProjectService struct {
    db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
    return &ProjectService{db: db}
}

func (s *ProjectService) CreateProject(input *model.ProjectCreateInput) (*model.Project, error) {
    project := &model.Project{
        UserID:          input.UserID,
        ProjectTitle:    input.ProjectTitle,
        Category:        input.Category,
        Description:     input.Description,
        Skills:          input.Skills,
        ContactName:     input.ContactName,
        ContactPhone:    input.ContactPhone,
        AdditionalNotes: input.AdditionalNotes,
        PlanTitle:       input.PlanTitle,
        Status:          "pending",
    }

    result := s.db.Create(project)
    if result.Error != nil {
        return nil, result.Error
    }

    return project, nil
}

func (s *ProjectService) GetProjectByID(id uint) (*model.Project, error) {
    var project model.Project
    result := s.db.First(&project, id)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, errors.New("project not found")
        }
        return nil, result.Error
    }
    return &project, nil
}

func (s *ProjectService) GetProjectsByUserID(userID int) ([]model.Project, error) {
    var projects []model.Project
    result := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&projects)
    if result.Error != nil {
        return nil, result.Error
    }
    return projects, nil
}

func (s *ProjectService) GetAllProjects() ([]model.Project, error) {
    var projects []model.Project
    result := s.db.Where("deleted_at IS NULL").Order("created_at DESC").Find(&projects)
    if result.Error != nil {
        return nil, result.Error
    }
    return projects, nil
}

func (s *ProjectService) UpdateProjectStatus(id uint, status string) error {
    result := s.db.Model(&model.Project{}).Where("id = ?", id).Update("status", status)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("project not found")
    }
    return nil
}

func (s *ProjectService) DeleteProject(id uint) error {
    result := s.db.Delete(&model.Project{}, id)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("project not found")
    }
    return nil
}