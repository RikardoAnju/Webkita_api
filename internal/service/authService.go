package service

import (
	"errors"
	"log"
	"strings"
	"time"

	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"BackendFramework/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	GroupIDAdmin = 1
	GroupIDUser  = 2
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: database.DbWebkita,
	}
}

// =========================================================================
// REGISTER & LOGIN
// =========================================================================

func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	if req.Username == "" {
		return nil, errors.New("username wajib diisi")
	}
	if len(req.Username) < 3 {
		return nil, errors.New("username minimal 3 karakter")
	}
	if err := s.cekUsernameAda(req.Username); err != nil {
		return nil, err
	}
	if req.Email == "" {
		return nil, errors.New("email wajib diisi")
	}
	if !s.isValidEmail(req.Email) {
		return nil, errors.New("format email tidak valid")
	}
	if err := s.cekEmailAda(req.Email); err != nil {
		return nil, err
	}
	if req.Phone != "" {
		if err := s.cekTeleponAda(req.Phone); err != nil {
			return nil, err
		}
	}
	if req.Password == "" {
		return nil, errors.New("password wajib diisi")
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}
	if req.ConfirmPassword == "" {
		return nil, errors.New("konfirmasi password wajib diisi")
	}
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("password dan konfirmasi password tidak sama")
	}
	if req.GroupID == 0 {
		req.GroupID = GroupIDUser
	}
	if req.IsAktif == "" {
		req.IsAktif = "Y"
	}
	if !s.isValidGroupID(req.GroupID) {
		return nil, errors.New("ID grup tidak valid (harus 1 atau 2)")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, errors.New("gagal mengenkripsi kata sandi")
	}

	now := time.Now()
	user := model.User{
		Username:            req.Username,
		Email:               req.Email,
		Password:            string(hashedPassword),
		GroupID:             req.GroupID,
		IsAktif:             req.IsAktif,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Phone:               req.Phone,
		SubscribeNewsletter: req.SubscribeNewsletter,
		CreatedAt:           &now,
		UpdatedAt:           &now,
	}

	if err := s.db.Create(&user).Error; err != nil {
		log.Printf("Error insert user: %v", err)
		return nil, errors.New("gagal membuat akun pengguna")
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *AuthService) Login(req model.LoginRequest) (*model.UserResponse, string, error) {
	user, err := s.cariUserAktifBerdasarkanEmail(req.Email)
	if err != nil {
		return nil, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", errors.New("email atau kata sandi salah")
	}
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, "", errors.New("gagal membuat token otentikasi")
	}
	response := user.ToResponse()
	return &response, token, nil
}

func (s *AuthService) LoginWithUsername(req model.LoginWithUsernameRequest) (*model.UserResponse, string, error) {
	user, err := s.cariUserAktifBerdasarkanUsername(req.Username)
	if err != nil {
		return nil, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", errors.New("username atau kata sandi salah")
	}
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, "", errors.New("gagal membuat token otentikasi")
	}
	response := user.ToResponse()
	return &response, token, nil
}

func (s *AuthService) GetUserByID(userID interface{}) (*model.UserResponse, error) {
	user, err := s.getUserBerdasarkanID(userID)
	if err != nil {
		return nil, err
	}
	response := user.ToResponse()
	return &response, nil
}

func (s *AuthService) GetOneUserByUsername(username string) *model.User {
	user, err := s.cariUserAktifBerdasarkanUsername(username)
	if err != nil {
		return nil
	}
	return user
}

// =========================================================================
// TOKEN
// =========================================================================

func (s *AuthService) UpsertTokenData(userID string, tokenData map[string]interface{}) bool {
	var existing model.UserToken
	err := s.db.Where("user_id = ?", userID).First(&existing).Error

	now := time.Now()
	tokenData["updated_at"] = now

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert baru
		token := model.UserToken{
			UserID:              userID,
			LastIPAddress:       toString(tokenData["last_ip_address"]),
			LastUserAgent:       toString(tokenData["last_user_agent"]),
			AccessToken:         toString(tokenData["access_token"]),
			RefreshToken:        toString(tokenData["refresh_token"]),
			RefreshTokenExpired: toTime(tokenData["refresh_token_expired"]),
			LastLogin:           toTime(tokenData["last_login"]),
			IsValidToken:        toString(tokenData["is_valid_token"]),
			IsRememberMe:        toString(tokenData["is_remember_me"]),
			CreatedAt:           now,
			UpdatedAt:           now,
		}
		if err := s.db.Create(&token).Error; err != nil {
			log.Printf("Error insert token: %v", err)
			return false
		}
	} else if err == nil {
		// Update existing
		updates := map[string]interface{}{
			"last_ip_address":       tokenData["last_ip_address"],
			"last_user_agent":       tokenData["last_user_agent"],
			"access_token":          tokenData["access_token"],
			"refresh_token":         tokenData["refresh_token"],
			"refresh_token_expired": tokenData["refresh_token_expired"],
			"last_login":            tokenData["last_login"],
			"is_valid_token":        tokenData["is_valid_token"],
			"is_remember_me":        tokenData["is_remember_me"],
			"updated_at":            now,
		}
		if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
			log.Printf("Error update token: %v", err)
			return false
		}
	} else {
		log.Printf("Error checking token: %v", err)
		return false
	}

	return true
}

func (s *AuthService) GetTokenData(userID, refreshToken string) map[string]interface{} {
	var token model.UserToken
	err := s.db.Where("user_id = ? AND refresh_token = ? AND is_valid_token = 'Y'", userID, refreshToken).
		First(&token).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Error getting token: %v", err)
		}
		return nil
	}

	return map[string]interface{}{
		"user_id":               token.UserID,
		"last_ip_address":       token.LastIPAddress,
		"last_user_agent":       token.LastUserAgent,
		"access_token":          token.AccessToken,
		"refresh_token":         token.RefreshToken,
		"refresh_token_expired": token.RefreshTokenExpired,
		"last_login":            token.LastLogin,
		"is_valid_token":        token.IsValidToken,
		"is_remember_me":        token.IsRememberMe,
		"created_at":            token.CreatedAt,
		"updated_at":            token.UpdatedAt,
	}
}

func (s *AuthService) DeleteTokenData(userID string) bool {
	if err := s.db.Where("user_id = ?", userID).Delete(&model.UserToken{}).Error; err != nil {
		log.Printf("Error deleting token: %v", err)
		return false
	}
	return true
}

func (s *AuthService) TestPing() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// =========================================================================
// HELPERS
// =========================================================================

func (s *AuthService) cekUsernameAda(username string) error {
	var count int64
	s.db.Model(&model.User{}).Where("username = ? AND deleted_at IS NULL", username).Count(&count)
	if count > 0 {
		return errors.New("username sudah terdaftar")
	}
	return nil
}

func (s *AuthService) cekEmailAda(email string) error {
	var count int64
	s.db.Model(&model.User{}).Where("email = ? AND deleted_at IS NULL", email).Count(&count)
	if count > 0 {
		return errors.New("email sudah terdaftar")
	}
	return nil
}

func (s *AuthService) cekTeleponAda(phone string) error {
	var count int64
	s.db.Model(&model.User{}).Where("phone = ? AND deleted_at IS NULL", phone).Count(&count)
	if count > 0 {
		return errors.New("nomor telepon sudah terdaftar")
	}
	return nil
}

func (s *AuthService) isValidGroupID(groupID uint) bool {
	return groupID == GroupIDAdmin || groupID == GroupIDUser
}

func (s *AuthService) isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	atIndex := strings.Index(email, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}
	dotIndex := strings.LastIndex(email, ".")
	if dotIndex == -1 || dotIndex < atIndex || dotIndex == len(email)-1 {
		return false
	}
	return true
}

func (s *AuthService) cariUserAktifBerdasarkanEmail(email string) (*model.User, error) {
	var user model.User
	err := s.db.Where("email = ? AND is_aktif = 'Y' AND deleted_at IS NULL", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if err != nil {
		return nil, errors.New("gagal memuat data pengguna")
	}
	return &user, nil
}

func (s *AuthService) cariUserAktifBerdasarkanUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.Where("username = ? AND is_aktif = 'Y' AND deleted_at IS NULL", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if err != nil {
		return nil, errors.New("gagal memuat data pengguna")
	}
	return &user, nil
}

func (s *AuthService) getUserBerdasarkanID(userID interface{}) (*model.User, error) {
	var user model.User
	err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if err != nil {
		return nil, errors.New("gagal memuat data pengguna")
	}
	return &user, nil
}

// =========================================================================
// UTILITY
// =========================================================================

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toTime(v interface{}) time.Time {
	if v == nil {
		return time.Time{}
	}
	t, _ := v.(time.Time)
	return t
}

// =========================================================================
// PACKAGE LEVEL (BACKWARD COMPAT)
// =========================================================================

var defaultAuthService *AuthService

func init() {
	defaultAuthService = NewAuthService()
}

func GetOneUserByUsername(username string) *model.User {
	return defaultAuthService.GetOneUserByUsername(username)
}

func UpsertTokenData(userID string, tokenData map[string]interface{}) bool {
	return defaultAuthService.UpsertTokenData(userID, tokenData)
}

func GetTokenData(userID, refreshToken string) map[string]interface{} {
	return defaultAuthService.GetTokenData(userID, refreshToken)
}

func DeleteTokenData(userID string) bool {
	return defaultAuthService.DeleteTokenData(userID)
}

func TestPing() error {
	return defaultAuthService.TestPing()
}