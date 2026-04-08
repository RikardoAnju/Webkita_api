package controller

import(
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"BackendFramework/internal/config"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/service"
	"BackendFramework/internal/model"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService: service.NewAuthService(),
	}
}

// Register handles user registration
func (ctrl *AuthController) Register(c *gin.Context) {
	var req model.RegisterRequest
	
	// Bind dan validasi JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📝 Register request - Username: %s, Email: %s, Role: %s", req.Username, req.Email, req.Role)

	
	user, err := ctrl.authService.Register(req)
	if err != nil {
		log.Printf("❌ Register failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// ✅ SUKSES: Response dengan status 201 Created
	log.Printf("✅ Register success - User ID: %d, Username: %s", user.ID, user.Username)
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registrasi berhasil",
		"data":    user,
	})
}

// LoginWithEmail handles email-based login
func (ctrl *AuthController) LoginWithEmail(c *gin.Context) {
	var req model.LoginRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📝 Login attempt - Email: %s", req.Email)

	user, token, err := ctrl.authService.Login(req)
	if err != nil {
		log.Printf("❌ Login failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	log.Printf("✅ Login success - User ID: %d, Email: %s", user.ID, user.Email)
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Login berhasil",
		"user":        user,
		"accessToken": token,
	})
}


func (ctrl *AuthController) LoginWithUsername(c *gin.Context) {
	var req model.LoginWithUsernameRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📝 Login attempt - Username: %s", req.Username)

	user, token, err := ctrl.authService.LoginWithUsername(req)
	if err != nil {
		log.Printf("❌ Login failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	log.Printf("✅ Login success - User ID: %d, Username: %s", user.ID, user.Username)
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Login berhasil",
		"user":        user,
		"accessToken": token,
	})
}

// Login handles LDAP-based login (legacy compatibility)
func Login(c *gin.Context) {
	var loginBody struct {
		Email      string `json:"email" binding:"required"`
		Password   string `json:"password" binding:"required"`
		RememberMe string `json:"remember_me"`
	}
	
	if err := c.ShouldBindJSON(&loginBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
		})
		return
	}
	
	log.Printf("📝 LDAP Login attempt - Email: %s", loginBody.Email)
	
	// LDAP Authentication
	ldapValid, err := ldapAuth(loginBody.Email, loginBody.Password)
	if !ldapValid {
		log.Printf("❌ LDAP authentication failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Verifikasi LDAP gagal",
		})
		return
	}
	
	// Get user from database
	user := service.GetOneUserByUsername(loginBody.Email)
	if user == nil {
		log.Printf("❌ User not found: %s", loginBody.Email)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Pengguna tidak ditemukan",
		})
		return
	}

	// Generate tokens
	accessToken, err := middleware.GenerateAccessToken(user.Username)
	if err != nil {
		log.Printf("❌ Failed to generate access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal membuat token akses",
		})
		return
	}
	
	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		log.Printf("❌ Failed to generate refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal membuat refresh token",
		})
		return
	}
	
	// Save token data
	tokenData := map[string]interface{}{
		"last_ip_address":         c.ClientIP(),
		"last_user_agent":         c.GetHeader("User-Agent"),
		"access_token":            accessToken,
		"refresh_token":           refreshToken,
		"refresh_token_expired":   time.Now().Add(config.RefreshTokenExpiry),
		"last_login":              time.Now(),
		"is_valid_token":          "y",
		"is_remember_me":          loginBody.RememberMe,
	}
	
	if !service.UpsertTokenData(user.Username, tokenData) {
		log.Printf("❌ Failed to save token data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menyimpan token ke database",
		})
		return
	}
	
	log.Printf("✅ LDAP Login success - Username: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"message":      "Login berhasil",
		"userId":       user.Username,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	userID := c.Param("usrId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID tidak diberikan",
		})
		return
	}
	
	log.Printf("📝 Logout request - User ID: %s", userID)
	
	if !service.DeleteTokenData(userID) {
		log.Printf("❌ Failed to delete token data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal logout pengguna",
		})
		return
	}
	
	log.Printf("✅ Logout success - User ID: %s", userID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logout berhasil",
	})
}

// RefreshAccessToken handles token refresh
func RefreshAccessToken(c *gin.Context) {
	var refreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
		UserID       string `json:"user_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&refreshRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Refresh token dan user ID wajib diisi",
		})
		return
	}
	
	log.Printf("📝 Token refresh request - User ID: %s", refreshRequest.UserID)
	
	// Get stored token data
	storedToken := service.GetTokenData(refreshRequest.UserID, refreshRequest.RefreshToken)
	
	if storedToken == nil {
		log.Printf("❌ Invalid refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Refresh token tidak valid",
		})
		return
	}
	
	// Extract values from map
	userID, ok := storedToken["user_id"].(string)
	if !ok {
		log.Printf("❌ Invalid token data format")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Format data token tidak valid",
		})
		return
	}
	
	// Check if refresh token is expired
	extendRefresh := false
	if refreshTokenExpired, ok := storedToken["refresh_token_expired"].(time.Time); ok {
		if time.Now().After(refreshTokenExpired) {
			isRememberMe, _ := storedToken["is_remember_me"].(string)
			if isRememberMe != "y" {
				log.Printf("❌ Refresh token expired")
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  "error",
					"message": "Refresh token kadaluarsa",
				})
				return
			}
			extendRefresh = true
		}
	}
	
	// Generate new access token
	newAccessToken, err := middleware.GenerateAccessToken(userID)
	if err != nil {
		log.Printf("❌ Failed to generate new access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal membuat access token baru",
		})
		return
	}

	// Prepare update data
	updateData := map[string]interface{}{
		"access_token": newAccessToken,
	}
	
	// Extend refresh token if remember me is enabled
	if extendRefresh {
		updateData["refresh_token_expired"] = time.Now().Add(config.RefreshTokenExpiry)
	}

	// Update token data in database
	if !service.UpsertTokenData(userID, updateData) {
		log.Printf("❌ Failed to update token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memperbarui token",
		})
		return
	}

	log.Printf("✅ Token refresh success - User ID: %s", userID)
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Token berhasil diperbarui",
		"accessToken": newAccessToken,
	})
}

// GetProfile handles getting user profile
func (ctrl *AuthController) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Pengguna tidak terautentikasi",
		})
		return
	}

	log.Printf("📝 Get profile request - User ID: %s", userID)

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		log.Printf("❌ User not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	log.Printf("✅ Profile retrieved - User ID: %s", userID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profil berhasil diambil",
		"data":    user,
	})
}

// ldapAuth performs LDAP authentication
func ldapAuth(username, password string) (bool, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.LDAP_SERVER, config.LDAP_PORT))
	if err != nil {
		middleware.LogError(err, "Failed to dial to LDAP server")
		return false, err
	}
	defer l.Close()

	err = l.Bind(username, password)
	if err != nil {
		return false, err
	}
	
	return true, nil
}