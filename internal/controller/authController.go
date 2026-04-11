package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		authService: service.NewAuthService(db),
	}
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func respond(c *gin.Context, status int, message string, data interface{}) {
	body := gin.H{"status": statusText(status), "message": message}
	if data != nil {
		body["data"] = data
	}
	c.JSON(status, body)
}

func respondError(c *gin.Context, err error) {
	var appErr *model.AppError
	if errors.As(err, &appErr) {
		code := appErr.Code
		if code == 0 {
			code = http.StatusBadRequest
		}
		c.JSON(code, gin.H{
			"status":  "error",
			"message": appErr.Message,
			"field":   appErr.Field,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  "error",
		"message": "Terjadi kesalahan internal server",
	})
}

func statusText(code int) string {
	if code >= 200 && code < 300 {
		return "success"
	}
	return "error"
}

// ─── POST /auth/register ──────────────────────────────────────────────────────

func (ctrl *AuthController) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Register bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📝 Register request: %s (%s)", req.Username, req.Email)

	user, err := ctrl.authService.Register(req)
	if err != nil {
		log.Printf("❌ Register failed: %v", err)
		respondError(c, err)
		return
	}

	log.Printf("✅ Register sukses: ID=%d %s", user.ID, user.Username)
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registrasi berhasil! Silakan cek email Anda untuk verifikasi akun.",
		"data":    user,
	})
}

// ─── GET /auth/verify-email ───────────────────────────────────────────────────

func (ctrl *AuthController) VerifyEmail(c *gin.Context) {
	var req model.VerifyEmailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Parameter token dan email wajib diisi",
		})
		return
	}

	log.Printf("📝 Verify email: %s", req.Email)

	if err := ctrl.authService.VerifyEmail(req); err != nil {
		log.Printf("❌ Verify email failed: %v", err)
		respondError(c, err)
		return
	}

	log.Printf("✅ Email verified: %s", req.Email)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Email berhasil diverifikasi! Akun Anda sudah aktif.",
	})
}

// ─── POST /auth/resend-verification ──────────────────────────────────────────

func (ctrl *AuthController) ResendVerification(c *gin.Context) {
	var req model.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format email tidak valid",
		})
		return
	}

	log.Printf("📝 Resend verification: %s", req.Email)

	if err := ctrl.authService.ResendVerification(req); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Jika email Anda terdaftar dan belum diverifikasi, email baru akan segera dikirim.",
	})
}

// ─── POST /auth/login ─────────────────────────────────────────────────────────

func (ctrl *AuthController) LoginWithEmail(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📝 Login (email): %s", req.Email)

	user, token, err := ctrl.authService.Login(req)
	if err != nil {
		log.Printf("❌ Login failed: %v", err)
		respondError(c, err)
		return
	}

	log.Printf("✅ Login sukses: %s", user.Email)
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Login berhasil",
		"user":        user,
		"accessToken": token,
	})
}

// ─── POST /auth/login-username ────────────────────────────────────────────────

func (ctrl *AuthController) LoginWithUsername(c *gin.Context) {
	var req model.LoginWithUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📝 Login (username): %s", req.Username)

	user, token, err := ctrl.authService.LoginWithUsername(req)
	if err != nil {
		log.Printf("❌ Login failed: %v", err)
		respondError(c, err)
		return
	}

	log.Printf("✅ Login sukses: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Login berhasil",
		"user":        user,
		"accessToken": token,
	})
}

// ─── GET /auth/profile ────────────────────────────────────────────────────────

func (ctrl *AuthController) GetProfile(c *gin.Context) {
	rawID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Pengguna tidak terautentikasi",
		})
		return
	}

	var userID uint
	switch v := rawID.(type) {
	case uint:
		userID = v
	case float64:
		userID = uint(v)
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User ID tidak valid"})
			return
		}
		userID = uint(id)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User ID tidak valid"})
		return
	}

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profil berhasil diambil",
		"data":    user,
	})
}

// ─── POST /auth/forgot-password ──────────────────────────────────────────────

func (ctrl *AuthController) ForgotPassword(c *gin.Context) {
	var req model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format email tidak valid",
		})
		return
	}

	log.Printf("📝 Forgot password: %s", req.Email)

	otpToken, err := ctrl.authService.ForgotPassword(req)
	if err != nil {
		log.Printf("❌ Forgot password error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Terjadi kesalahan internal server",
		})
		return
	}

	// Selalu 200 — jangan bocorkan apakah email terdaftar
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Jika email terdaftar, kode OTP akan dikirim dalam beberapa saat.",
		"otp_token": otpToken, // FE simpan ini, kirim balik saat reset-password
	})
}

// ─── POST /auth/reset-password ────────────────────────────────────────────────

func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📝 Reset password request")

	if err := ctrl.authService.ResetPassword(req); err != nil {
		log.Printf("❌ Reset password failed: %v", err)
		respondError(c, err)
		return
	}

	log.Printf("✅ Reset password sukses")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Password berhasil direset. Silakan login dengan password baru.",
	})
}