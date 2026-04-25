package middleware

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userIDRaw.(string)
		if !ok || userIDStr == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid user ID",
			})
			c.Abort()
			return
		}
		var user model.User
		err := database.DbWebkita.
			Where("id = ? AND deleted_at IS NULL AND is_aktif = ?", userIDStr, "Y").
			First(&user).Error

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "User not found",
			})
			c.Abort()
			return
		}

		if user.GroupID != 1 {
			c.JSON(http.StatusOK, gin.H{
				"code":  http.StatusForbidden,
				"error": "Access denied. Admin only.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
