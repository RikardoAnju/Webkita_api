package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"BackendFramework/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	verificationTokenExpiry = 24 * time.Hour
	bcryptCost              = 12
)

type AuthService struct {
	db           *gorm.DB
	emailService *EmailService
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db:           db,
		emailService: NewEmailService(),
	}
}

// ─── Register ─────────────────────────────────────────────────────────────────

// Register membuat akun baru dan mengirim email verifikasi secara async
func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	// Cek duplikat email
	var count int64
	if err := s.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("gagal cek email: %w", err)
	}
	if count > 0 {
		return nil, model.ErrEmailExists
	}

	// Cek duplikat username
	if err := s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("gagal cek username: %w", err)
	}
	if count > 0 {
		return nil, model.ErrUsernameExists
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("gagal hash password: %w", err)
	}

	// Generate token verifikasi
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("gagal buat token: %w", err)
	}
	expiry := time.Now().Add(verificationTokenExpiry)

	groupID := req.GroupID
	if groupID == 0 {
		groupID = 2
	}

	user := model.User{
		Username:            req.Username,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               req.Email,
		Phone:               req.Phone,
		Password:            string(hashed),
		GroupID:             groupID,
		Role:                model.RoleUser,
		IsAktif:             model.StatusInaktif, // aktif setelah verifikasi
		SubscribeNewsletter: req.SubscribeNewsletter,
		VerificationToken:   token,
		TokenExpiresAt:      &expiry,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("gagal simpan user: %w", err)
	}

	// Kirim email secara async (tidak block response)
	go func() {
		if err := s.emailService.SendVerificationEmail(user.Email, user.Username, token); err != nil {
			log.Printf("⚠️  Gagal kirim email verifikasi ke %s: %v", user.Email, err)
		}
	}()

	log.Printf("✅ Register berhasil: %s (%s)", user.Username, user.Email)
	resp := user.ToResponse()
	return &resp, nil
}

// ─── Verify Email ─────────────────────────────────────────────────────────────

// VerifyEmail memvalidasi token dan mengaktifkan akun
func (s *AuthService) VerifyEmail(req model.VerifyEmailRequest) error {
	var user model.User
	err := s.db.Where("email = ? AND verification_token = ?", req.Email, req.Token).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ErrInvalidToken
		}
		return fmt.Errorf("gagal query user: %w", err)
	}

	if user.IsEmailVerified() {
		return model.ErrEmailAlreadyVerified
	}

	if user.IsTokenExpired() {
		return model.ErrInvalidToken
	}

	now := time.Now()
	err = s.db.Model(&user).Updates(map[string]interface{}{
		"email_verified_at":  now,
		"is_aktif":           model.StatusAktif,
		"verification_token": "",
		"token_expires_at":   nil,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal aktivasi akun: %w", err)
	}

	log.Printf("✅ Email verified: %s (%s)", user.Username, user.Email)
	return nil
}

// ─── Resend Verification ──────────────────────────────────────────────────────

// ResendVerification membuat token baru dan kirim ulang email verifikasi
func (s *AuthService) ResendVerification(req model.ResendVerificationRequest) error {
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Kembalikan sukses walau tidak ditemukan (keamanan: jangan bocorkan info email)
			return nil
		}
		return fmt.Errorf("gagal query user: %w", err)
	}

	if user.IsEmailVerified() {
		return model.ErrEmailAlreadyVerified
	}

	// Buat token baru
	token, err := generateSecureToken(32)
	if err != nil {
		return fmt.Errorf("gagal buat token: %w", err)
	}
	expiry := time.Now().Add(verificationTokenExpiry)

	err = s.db.Model(&user).Updates(map[string]interface{}{
		"verification_token": token,
		"token_expires_at":   expiry,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal update token: %w", err)
	}

	go func() {
		if err := s.emailService.SendResendVerificationEmail(user.Email, user.Username, token); err != nil {
			log.Printf("⚠️  Gagal kirim ulang email ke %s: %v", user.Email, err)
		}
	}()

	log.Printf("📧 Kirim ulang verifikasi ke: %s", user.Email)
	return nil
}

// ─── Login (Email) ────────────────────────────────────────────────────────────

func (s *AuthService) Login(req model.LoginRequest) (*model.UserResponse, string, error) {
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", model.ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", model.ErrInvalidCredentials
	}

	if !user.IsEmailVerified() {
		return nil, "", model.ErrEmailNotVerified
	}

	if user.IsAktif != model.StatusAktif {
		return nil, "", model.ErrUserInactive
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	resp := user.ToResponse()
	return &resp, token, nil
}

// ─── Login (Username) ─────────────────────────────────────────────────────────

func (s *AuthService) LoginWithUsername(req model.LoginWithUsernameRequest) (*model.UserResponse, string, error) {
	var user model.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", model.ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", model.ErrInvalidCredentials
	}

	if !user.IsEmailVerified() {
		return nil, "", model.ErrEmailNotVerified
	}

	if user.IsAktif != model.StatusAktif {
		return nil, "", model.ErrUserInactive
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	resp := user.ToResponse()
	return &resp, token, nil
}

// ─── Get User By ID ───────────────────────────────────────────────────────────

func (s *AuthService) GetUserByID(id uint) (*model.UserResponse, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateJWT — sesuaikan dengan middleware JWT Anda
func generateJWT(user model.User) (string, error) {
	// TODO: Ganti dengan pemanggilan middleware.GenerateAccessToken(user.Username)
	// Contoh placeholder:
	return fmt.Sprintf("jwt_placeholder_for_%s", user.Username), nil
}