package controller

import (
	"net/http"

	"BackendFramework/internal/model"
	"BackendFramework/internal/service"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	usrId := c.Param("usrId")

	if usrId != "" {
		// Get single user
		user, err := service.GetOneUser(usrId)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "User not found",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "User retrieved successfully",
			"data":    user,
		})
		return
	}

	// Get all users
	users, err := service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to retrieve users",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Users retrieved successfully",
		"data":    users,
		"total":   len(users),
	})
}

// GetUserByEmail mengambil user berdasarkan email
func GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Email parameter is required",
		})
		return
	}

	user, err := service.GetOneUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "User not found",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "User retrieved successfully",
		"data":    user,
	})
}

// GET /auth/me atau /user/profile
func GetMyProfile(c *gin.Context) {
	// Ambil user ID dari JWT claims (sudah di-set oleh middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := service.GetCurrentUserProfile(userID.(uint))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status": "success",
		"data":   user.ToResponse(),
	})
}

// InsertUser menambahkan user baru
func InsertUser(c *gin.Context) {
	// Get validated input from middleware
	validatedInput, exists := c.Get("validatedInput")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request data",
		})
		return
	}

	userInput, ok := validatedInput.(*model.UserInput)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid user input format",
		})
		return
	}

	// Check if user already exists
	exists, err := service.CheckUserExists(userInput.Username, userInput.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to check user existence",
			"error":   err.Error(),
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"code":    http.StatusConflict,
			"message": "User with this username or email already exists",
		})
		return
	}

	// Insert user
	err = service.InsertUser(userInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	// Don't return password in response
	userInput.Password = ""

	c.JSON(http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": "User created successfully",
		"data":    userInput,
	})
}

// UpdateUser memperbarui data user
func UpdateUser(c *gin.Context) {
	// Get validated input from middleware
	validatedInput, exists := c.Get("validatedInput")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request data",
		})
		return
	}

	userInput, ok := validatedInput.(*model.UserInput)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid user input format",
		})
		return
	}

	// Update user
	err := service.UpdateUser(userInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to update user",
			"error":   err.Error(),
		})
		return
	}

	// Don't return password in response
	userInput.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "User updated successfully",
		"data":    userInput,
	})
}

// DeleteUser menghapus user (soft delete)
func DeleteUser(c *gin.Context) {
	usrId := c.Param("usrId")
	if usrId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "User ID is required",
		})
		return
	}

	err := service.DeleteUser(usrId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to delete user",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "User deleted successfully",
		"data": gin.H{
			"username": usrId,
		},
	})
}

// HardDeleteUser menghapus user secara permanen
func HardDeleteUser(c *gin.Context) {
	usrId := c.Param("usrId")
	if usrId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "User ID is required",
		})
		return
	}

	err := service.HardDeleteUser(usrId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to permanently delete user",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "User permanently deleted",
		"data": gin.H{
			"username": usrId,
		},
	})
}

func RestoreUser(c *gin.Context) {
	usrId := c.Param("usrId")
	if usrId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "User ID is required",
		})
		return
	}

	err := service.RestoreUser(usrId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to restore user",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "User restored successfully",
		"data": gin.H{
			"username": usrId,
		},
	})
}

func GetUserProfile(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "User not authenticated",
		})
		return
	}

	role, _ := c.Get("role")
	username, _ := c.Get("username")

	// Get full user details
	user, err := service.GetOneUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "User profile not found",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Profile retrieved successfully",
		"data": gin.H{
			"userID":   userID,
			"username": username,
			"role":     role,
			"profile":  user,
		},
	})
}

// UpdateUserPassword memperbarui password user
func UpdateUserPassword(c *gin.Context) {
	var input struct {
		Username    string `json:"username" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	err := service.UpdateUserPassword(input.Username, input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to update password",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Password updated successfully",
	})
}
