package model

import (
	"time"
	"gorm.io/gorm"
)

type UserList struct {
	Username string
	Email    string
	GroupID  uint
	IsAktif  string
	Password string
}

type UserInput struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	GroupID  uint   `json:"group_id" validate:"required"`
	IsAktif  string `json:"is_aktif"`
	Password string `json:"password" validate:"required,min=8"`
}

type FileInput struct {
	FileUpload      string `json:"fileusername" validate:"required"`
	FileDescription string `json:"filedescription" validate:"required"`
}


type User struct {
	ID                  uint            `json:"id" gorm:"primaryKey"`
	Username            string          `json:"username" gorm:"uniqueIndex;not null"`
	FirstName           string          `json:"first_name"`
	LastName            string          `json:"last_name"`
	Email               string          `json:"email" gorm:"uniqueIndex;not null"`
	Phone               string          `json:"phone" gorm:"index"`
	Password            string          `json:"-" gorm:"not null"`
	GroupID             uint            `json:"group_id" gorm:"default:2;not null;index"`
	Role                string          `json:"role" gorm:"default:pembeli;size:50"` 
	IsAktif             string          `json:"is_aktif" gorm:"default:Y;size:1;not null"`
	SubscribeNewsletter bool            `json:"subscribe_newsletter" gorm:"default:false"`
	CreatedAt           *time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt           *time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt           *gorm.DeletedAt `json:"-" gorm:"index"`
}


type RegisterRequest struct {
	Username            string `json:"username" binding:"required,min=3"`
	Email               string `json:"email" binding:"required,email"`
	Password            string `json:"password" binding:"required,min=8"`
	ConfirmPassword     string `json:"confirmPassword,omitempty"`
	Role                string `json:"role"` 
	GroupID             uint   `json:"group_id"`
	IsAktif             string `json:"is_aktif" binding:"omitempty,oneof=Y N"`
	FirstName           string `json:"first_name"`
	LastName            string `json:"last_name"`
	Phone               string `json:"phone"`
	SubscribeNewsletter bool   `json:"subscribe_newsletter"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginWithUsernameRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}


type UserResponse struct {
	ID                  uint       `json:"id"`
	Username            string     `json:"username"`
	FirstName           string     `json:"first_name,omitempty"`
	LastName            string     `json:"last_name,omitempty"`
	Email               string     `json:"email"`
	Phone               string     `json:"phone,omitempty"`
	GroupID             uint       `json:"group_id"`
	Role                string     `json:"role,omitempty"`
	IsAktif             string     `json:"is_aktif"`
	SubscribeNewsletter bool       `json:"subscribe_newsletter"`
	CreatedAt           *time.Time `json:"createdAt"`
}


func (User) TableName() string {
	return "users"
}


func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                  u.ID,
		Username:            u.Username,
		FirstName:           u.FirstName,
		LastName:            u.LastName,
		Email:               u.Email,
		Phone:               u.Phone,
		GroupID:             u.GroupID,
		Role:                u.Role,
		IsAktif:             u.IsAktif,
		SubscribeNewsletter: u.SubscribeNewsletter,
		CreatedAt:           u.CreatedAt,
	}
}

func (u *User) ToUserList() UserList {
	return UserList{
		Username: u.Username,
		Email:    u.Email,
		GroupID:  u.GroupID,
		IsAktif:  u.IsAktif,
		Password: "",
	}
}

func (ui *UserInput) ToUser() User {
	isAktif := ui.IsAktif
	if isAktif == "" {
		isAktif = "Y"
	}
	if isAktif != "Y" && isAktif != "N" {
		isAktif = "Y"
	}

	groupID := ui.GroupID
	if groupID == 0 {
		groupID = 2
	}

	return User{
		Username: ui.Username,
		Email:    ui.Email,
		GroupID:  groupID,
		IsAktif:  isAktif,
		Password: ui.Password,
	}
}

func (r *RegisterRequest) ToUser() User {
	isAktif := r.IsAktif
	if isAktif == "" {
		isAktif = "Y"
	}
	if isAktif != "Y" && isAktif != "N" {
		isAktif = "Y"
	}

	groupID := r.GroupID
	if groupID == 0 {
		groupID = 2
	}

	
	role := r.Role
	if role == "" {
		role = "pembeli"
	}
	
	if role != "pembeli" && role != "developer" {
		role = "pembeli"
	}

	return User{
		Username:            r.Username,
		Email:               r.Email,
		Password:            r.Password,
		FirstName:           r.FirstName,
		LastName:            r.LastName,
		Phone:               r.Phone,
		GroupID:             groupID,
		Role:                role,
		IsAktif:             isAktif,
		SubscribeNewsletter: r.SubscribeNewsletter,
	}
}

// =======================
// 🧩 Validation
// =======================
func (r *RegisterRequest) Validate() error {
	if r.Username == "" {
		return ErrUsernameRequired
	}
	if len(r.Username) < 3 {
		return ErrUsernameMinLength
	}
	if r.Email == "" {
		return ErrEmailRequired
	}
	if r.Password == "" {
		return ErrPasswordRequired
	}
	if len(r.Password) < 8 {
		return ErrPasswordMinLength
	}
	if r.ConfirmPassword != "" && r.Password != r.ConfirmPassword {
		return ErrPasswordMismatch
	}
	if r.IsAktif != "" && r.IsAktif != "Y" && r.IsAktif != "N" {
		return ErrInvalidIsAktif
	}
	if r.GroupID != 0 && r.GroupID < 1 {
		return ErrInvalidGroupID
	}
	if r.Role != "" && r.Role != "pembeli" && r.Role != "developer" {
		return ErrInvalidRole
	}
	return nil
}


var (
	ErrUsernameRequired  = &ValidationError{Field: "username", Message: "Username wajib diisi"}
	ErrUsernameMinLength = &ValidationError{Field: "username", Message: "Username minimal 3 karakter"}
	ErrEmailRequired     = &ValidationError{Field: "email", Message: "Email wajib diisi"}
	ErrPasswordRequired  = &ValidationError{Field: "password", Message: "Password wajib diisi"}
	ErrPasswordMinLength = &ValidationError{Field: "password", Message: "Password minimal 8 karakter"}
	ErrPasswordMismatch  = &ValidationError{Field: "confirmPassword", Message: "Password dan konfirmasi password tidak sama"}
	ErrInvalidIsAktif    = &ValidationError{Field: "is_aktif", Message: "IsAktif hanya boleh Y atau N"}
	ErrInvalidGroupID    = &ValidationError{Field: "group_id", Message: "ID Grup tidak valid"}
	ErrInvalidRole       = &ValidationError{Field: "role", Message: "Role hanya boleh pembeli atau developer"}
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}