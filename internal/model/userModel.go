package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"

	StatusAktif   = "Y"
	StatusInaktif = "N"
)

// ─── DB Model ─────────────────────────────────────────────────────────────────

type User struct {
	ID                  uint   `json:"id" gorm:"primaryKey"`
	Username            string `json:"username" gorm:"uniqueIndex;not null;size:100"`
	FirstName           string `json:"first_name" gorm:"size:100"`
	LastName            string `json:"last_name" gorm:"size:100"`
	Email               string `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Phone               string `json:"phone" gorm:"index;size:20"`
	Password            string `json:"-" gorm:"not null"`
	GroupID             uint   `json:"group_id" gorm:"default:2;not null;index"`
	Role                string `json:"role" gorm:"default:user;size:50"`
	IsAktif             string `json:"is_aktif" gorm:"default:N;size:1;not null"`
	SubscribeNewsletter bool   `json:"subscribe_newsletter" gorm:"default:false"`

	// Email Verification
	EmailVerifiedAt   *time.Time `json:"email_verified_at" gorm:"index"`
	VerificationToken string     `json:"-" gorm:"size:100;index"`
	TokenExpiresAt    *time.Time `json:"-"`

	CreatedAt *time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt *time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt *gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string { return "users" }

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) IsTokenExpired() bool {
	if u.TokenExpiresAt == nil {
		return true
	}
	return time.Now().After(*u.TokenExpiresAt)
}

func (u *UserInput) ToUser() User {
	return User{
		Username:  u.Username,
		Email:     u.Email,
		GroupID:   u.GroupID,
		IsAktif:   u.IsAktif,
		Password:  u.Password,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Role:      u.Role,
	}
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
		IsEmailVerified:     u.IsEmailVerified(),
		SubscribeNewsletter: u.SubscribeNewsletter,
		CreatedAt:           u.CreatedAt,
	}
}

// ─── Request Types ────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Username            string `json:"username" binding:"required,min=3,max=50,alphanum"`
	FirstName           string `json:"first_name" binding:"omitempty,max=100"`
	LastName            string `json:"last_name" binding:"omitempty,max=100"`
	Email               string `json:"email" binding:"required,email,max=255"`
	Phone               string `json:"phone" binding:"omitempty,max=20"`
	Password            string `json:"password" binding:"required,min=8,max=72"`
	GroupID             uint   `json:"group_id"`
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

type VerifyEmailRequest struct {
	Token string `form:"token" binding:"required"`
	Email string `form:"email" binding:"required,email"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UserInput struct {
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	GroupID   uint   `json:"group_id" validate:"required"`
	IsAktif   string `json:"is_aktif"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
}

type UserList struct {
	Username string
	Email    string
	GroupID  uint
	IsAktif  string
	Password string
}

// ─── Reset Password Request Types ────────────────────────────────────────────

// ForgotPasswordRequest - step 1: minta OTP dikirim ke email
// Response: { otp_token: "<jwt>" }
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest - step 2: kirim otp_token (JWT) + OTP dari email + password baru
type ResetPasswordRequest struct {
	OTPToken        string `json:"otp_token"        binding:"required"` // JWT dari step 1
	OTP             string `json:"otp"              binding:"required,len=6"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
}

// ─── Response Types ───────────────────────────────────────────────────────────

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
	IsEmailVerified     bool       `json:"is_email_verified"`
	SubscribeNewsletter bool       `json:"subscribe_newsletter"`
	CreatedAt           *time.Time `json:"createdAt"`
}

// ─── Error Type ───────────────────────────────────────────────────────────────

type AppError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

// ─── Sentinel Errors ──────────────────────────────────────────────────────────

var (
	// 409 Conflict
	ErrEmailExists    = &AppError{Field: "email", Message: "Email sudah terdaftar", Code: 409}
	ErrUsernameExists = &AppError{Field: "username", Message: "Username sudah digunakan", Code: 409}

	// 400 Bad Request
	ErrInvalidToken         = &AppError{Field: "token", Message: "Token tidak valid atau sudah kadaluwarsa", Code: 400}
	ErrEmailAlreadyVerified = &AppError{Field: "email", Message: "Email sudah diverifikasi sebelumnya", Code: 400}
	ErrInvalidGroupID       = &AppError{Field: "group_id", Message: "ID Grup tidak valid", Code: 400}
	ErrInvalidIsAktif       = &AppError{Field: "is_aktif", Message: "IsAktif hanya boleh Y atau N", Code: 400}
	ErrPasswordMismatch     = &AppError{Field: "confirm_password", Message: "Password baru dan konfirmasi tidak cocok", Code: 400}
	ErrOTPInvalid           = &AppError{Field: "otp", Message: "Kode OTP tidak valid atau sudah kedaluwarsa", Code: 400}

	// 401 Unauthorized
	ErrInvalidCredentials = &AppError{Field: "credentials", Message: "Email/username atau password salah", Code: 401}

	// 403 Forbidden
	ErrEmailNotVerified = &AppError{Field: "email", Message: "Email belum diverifikasi, silakan cek inbox Anda", Code: 403}
	ErrUserInactive     = &AppError{Field: "user", Message: "Akun tidak aktif", Code: 403}

	// 404 Not Found
	ErrUserNotFound = &AppError{Field: "user", Message: "Pengguna tidak ditemukan", Code: 404}

	// 429 Too Many Requests
	ErrTooManyRequests = &AppError{Field: "request", Message: "Terlalu banyak permintaan, coba lagi dalam 1 menit", Code: 429}
)