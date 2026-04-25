package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	verificationTokenExpiry = 24 * time.Hour
	otpExpiry               = 5 * time.Minute
	bcryptCost              = 12
)

type otpClaims struct {
	Email   string `json:"email"`
	OTPHash string `json:"otp_hash"`
	jwt.RegisteredClaims
}

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

func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	var count int64
	if err := s.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("gagal cek email: %w", err)
	}
	if count > 0 {
		return nil, model.ErrEmailExists
	}

	if err := s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("gagal cek username: %w", err)
	}
	if count > 0 {
		return nil, model.ErrUsernameExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("gagal hash password: %w", err)
	}

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
		IsAktif:             model.StatusInaktif,
		SubscribeNewsletter: req.SubscribeNewsletter,
		VerificationToken:   token,
		TokenExpiresAt:      &expiry,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("gagal simpan user: %w", err)
	}

	if err := s.emailService.SendVerificationEmail(user.Email, user.Username, token); err != nil {
		log.Printf("⚠️  Gagal kirim email verifikasi ke %s: %v", user.Email, err)
	}

	log.Printf("✅ Register berhasil: %s (%s)", user.Username, user.Email)
	resp := user.ToResponse()
	return &resp, nil
}

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

func (s *AuthService) ResendVerification(req model.ResendVerificationRequest) error {
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("gagal query user: %w", err)
	}

	if user.IsEmailVerified() {
		return model.ErrEmailAlreadyVerified
	}

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

	if err := s.emailService.SendResendVerificationEmail(user.Email, user.Username, token); err != nil {
		log.Printf("⚠️  Gagal kirim ulang email ke %s: %v", user.Email, err)
	}

	log.Printf("📧 Kirim ulang verifikasi ke: %s", user.Email)
	return nil
}

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

func (s *AuthService) ForgotPassword(req model.ForgotPasswordRequest) (string, error) {
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("gagal query user: %w", err)
	}

	otp, err := generateOTP(6)
	if err != nil {
		return "", fmt.Errorf("gagal buat OTP: %w", err)
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("gagal hash OTP: %w", err)
	}

	claims := otpClaims{
		Email:   user.Email,
		OTPHash: string(hashedOTP),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(otpExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	otpToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := otpToken.SignedString([]byte(getJWTSecret()))
	if err != nil {
		return "", fmt.Errorf("gagal sign OTP token: %w", err)
	}

	if err := s.emailService.SendResetPasswordOTPEmail(user.Email, user.Username, otp); err != nil {
		log.Printf("⚠️  Gagal kirim OTP ke %s: %v", user.Email, err)
	}

	log.Printf("🔑 OTP reset password dikirim ke: %s", user.Email)
	return signedToken, nil
}

func (s *AuthService) ResetPassword(req model.ResetPasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return model.ErrPasswordMismatch
	}

	claims, err := parseOTPToken(req.OTPToken)
	if err != nil {
		return model.ErrOTPInvalid
	}

	if err := bcrypt.CompareHashAndPassword([]byte(claims.OTPHash), []byte(req.OTP)); err != nil {
		return model.ErrOTPInvalid
	}

	var user model.User
	if err := s.db.Where("email = ?", claims.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ErrOTPInvalid
		}
		return fmt.Errorf("gagal query user: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("gagal hash password: %w", err)
	}

	if err := s.db.Model(&user).Update("password", string(hashed)).Error; err != nil {
		return fmt.Errorf("gagal update password: %w", err)
	}

	log.Printf("✅ Password berhasil direset: %s (%s)", user.Username, user.Email)
	return nil
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateOTP(n int) (string, error) {
	otp := ""
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += num.String()
	}
	return otp, nil
}

func parseOTPToken(tokenStr string) (*otpClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &otpClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method tidak valid")
		}
		return []byte(getJWTSecret()), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token tidak valid")
	}

	claims, ok := token.Claims.(*otpClaims)
	if !ok {
		return nil, fmt.Errorf("claims tidak valid")
	}
	return claims, nil
}

func getJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "fallback_secret_ganti_ini"
}

func generateJWT(user model.User) (string, error) {
	userID := fmt.Sprintf("%d", user.ID)
	return middleware.GenerateAccessToken(userID)
}
