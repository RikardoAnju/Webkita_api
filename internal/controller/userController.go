package controller

import (
	"errors"
	"net/http"
	"strconv"

	"BackendFramework/internal/model"
	"BackendFramework/internal/service"

	"github.com/gin-gonic/gin"
)

// ─── Helper ───────────────────────────────────────────────────────────────────

func respondErr(c *gin.Context, err error) {
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

// getUserIDFromContext mengambil user_id dari JWT context (support uint & string)
func getUserIDFromContext(c *gin.Context) (uint, bool) {
	raw, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch v := raw.(type) {
	case uint:
		return v, true
	case float64:
		return uint(v), true
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(id), true
	}
	return 0, false
}

// ─── GET /user/:usrId  atau  GET /user ───────────────────────────────────────

func GetUser(c *gin.Context) {
	usrId := c.Param("usrId")

	if usrId != "" {
		user, err := service.GetOneUser(usrId)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "User tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "User berhasil diambil",
			"data":    user,
		})
		return
	}

	users, err := service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Users berhasil diambil",
		"data":    users,
		"total":   len(users),
	})
}

// ─── GET /user?email=xxx ──────────────────────────────────────────────────────

func GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Parameter email wajib diisi",
		})
		return
	}

	user, err := service.GetOneUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "User tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User berhasil diambil",
		"data":    user,
	})
}

// ─── GET /auth/me ─────────────────────────────────────────────────────────────

// GetMyProfile mengambil profil user yang sedang login (dari JWT)
func GetMyProfile(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Pengguna tidak terautentikasi",
		})
		return
	}

	user, err := service.GetCurrentUserProfile(userID)
	if err != nil {
		respondErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user.ToResponse(),
	})
}

// ─── POST /user ───────────────────────────────────────────────────────────────

func InsertUser(c *gin.Context) {
	validatedInput, exists := c.Get("validatedInput")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data request tidak valid",
		})
		return
	}

	userInput, ok := validatedInput.(*model.UserInput)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format input user tidak valid",
		})
		return
	}

	exists, err := service.CheckUserExists(userInput.Username, userInput.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memeriksa keberadaan user",
		})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "Username atau email sudah digunakan",
		})
		return
	}

	if err := service.InsertUser(userInput); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal membuat user",
		})
		return
	}

	userInput.Password = "" // jangan return password
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User berhasil dibuat",
		"data":    userInput,
	})
}

// ─── PUT /user ────────────────────────────────────────────────────────────────

func UpdateUser(c *gin.Context) {
	validatedInput, exists := c.Get("validatedInput")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data request tidak valid",
		})
		return
	}

	userInput, ok := validatedInput.(*model.UserInput)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format input user tidak valid",
		})
		return
	}

	if err := service.UpdateUser(userInput); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memperbarui user",
		})
		return
	}

	userInput.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User berhasil diperbarui",
		"data":    userInput,
	})
}

// ─── DELETE /user/:usrId (soft delete) ───────────────────────────────────────

func DeleteUser(c *gin.Context) {
	usrId := c.Param("usrId")
	if usrId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID wajib diisi",
		})
		return
	}

	if err := service.DeleteUser(usrId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menghapus user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User berhasil dihapus",
		"data":    gin.H{"username": usrId},
	})
}

// ─── DELETE /user/:usrId/permanent ───────────────────────────────────────────

func HardDeleteUser(c *gin.Context) {
	usrId := c.Param("usrId")
	if usrId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID wajib diisi",
		})
		return
	}

	if err := service.HardDeleteUser(usrId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menghapus user secara permanen",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User berhasil dihapus secara permanen",
		"data":    gin.H{"username": usrId},
	})
}

// ─── POST /user/:usrId/restore ────────────────────────────────────────────────

func RestoreUser(c *gin.Context) {
	usrId := c.Param("usrId")
	if usrId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID wajib diisi",
		})
		return
	}

	if err := service.RestoreUser(usrId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memulihkan user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User berhasil dipulihkan",
		"data":    gin.H{"username": usrId},
	})
}

// ─── GET /user/profile (dari JWT middleware) ──────────────────────────────────

// GetUserProfile — gunakan untuk endpoint yang pakai key "userID" (legacy middleware)
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID") // legacy key, beda dengan "user_id"
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Pengguna tidak terautentikasi",
		})
		return
	}

	role, _ := c.Get("role")
	username, _ := c.Get("username")

	user, err := service.GetOneUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Profil user tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profil berhasil diambil",
		"data": gin.H{
			"userID":   userID,
			"username": username,
			"role":     role,
			"profile":  user,
		},
	})
}

// ─── PUT /user/password ───────────────────────────────────────────────────────

func UpdateUserPassword(c *gin.Context) {
	var input struct {
		Username    string `json:"username" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	// TODO: Validasi OldPassword sebelum update
	// Contoh: service.VerifyPassword(input.Username, input.OldPassword)

	if err := service.UpdateUserPassword(input.Username, input.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memperbarui password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Password berhasil diperbarui",
	})
}