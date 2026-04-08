package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
)

func LogUserActivity() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			userID = c.PostForm("user_id")
		}

		queryParams := c.Request.URL.Query()

		var requestBody map[string]interface{}
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			json.Unmarshal(bodyBytes, &requestBody)
		}

		if requestBody == nil {
			formData := make(map[string]interface{})
			c.Request.ParseForm()
			for key, values := range c.Request.PostForm {
				if len(values) == 1 {
					formData[key] = values[0]
				} else {
					formData[key] = values
				}
			}
			requestBody = formData
		}

		queryParamsJSON, _ := json.Marshal(queryParams)
		requestBodyJSON, _ := json.Marshal(requestBody)

		activity := model.UserActivity{
			Userid:      toString(userID), // fix: UserID -> Userid
			Endpoint:    c.Request.URL.Path,
			Method:      c.Request.Method,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.GetHeader("User-Agent"),
			QueryParams: string(queryParamsJSON),
			RequestBody: string(requestBodyJSON),
			Timestamp:   time.Now(),
		}

		if err := database.DbWebkita.Create(&activity).Error; err != nil {
			c.JSON(200, gin.H{
				"code":  401,
				"error": "Failed To Log User Activity",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}