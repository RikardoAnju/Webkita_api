package controller

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"BackendFramework/internal/model"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/thirdparty"
)

func UploadFile(c *gin.Context) {
	validatedInput, _ := c.Get("validatedInput")
	userInput := validatedInput.(*model.FileInput)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusBadRequest,
			"error": "file is required",
		})
		return
	}

	const maxFileSize = 9 * 1024 * 1024 // Max file size in bytes (9 MB)
	// Allowed file extensions
	var allowedExtensions = []string{".pdf"}
	const uploadDir   = "./temp/" // Directory to save uploaded files locally

	// Check file type (by extension)
	ext := strings.ToLower(filepath.Ext(file.Filename))

	fileStatus,errMsg := middleware.ValidateFile(maxFileSize, file.Size,ext,allowedExtensions)
	if(fileStatus == false) {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusBadRequest,
			"error": errMsg,
		})
		return
	}

	// Save the file locally
	localFilePath := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, localFilePath); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusInternalServerError,
			"error": "failed to save file locally",
		})
		return
	}
	// Upload the file to S3
	s3Url, err := thirdparty.UploadFileBucket(localFilePath, "Akademik/test"+ext)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusInternalServerError,
			"error": "failed to upload file to S3",
		})
		return
	}

	// Delete the local file after successful upload
	if err := os.Remove(localFilePath); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusInternalServerError,
			"error": "failed to clean up local file",
		})
		return
	}

	// Return response with S3 URL and other data
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"message":     "file uploaded successfully",
		"s3_url":      s3Url,
		"userInput":	userInput,
	})
}