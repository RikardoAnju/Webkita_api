package controller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"BackendFramework/internal/middleware"
	"BackendFramework/internal/service"
	"BackendFramework/internal/thirdparty"

	"github.com/gin-gonic/gin"
)

// Fungsi pembantu lokal untuk mengkonversi GroupID menjadi string yang dapat dibaca
func getGroupNameFromID(groupID uint) string {
	switch groupID {
	case 1:
		return "Admin"
	case 2:
		return "User"
	default:
		return "Unknown"
	}
}

// TryGeneratePdf generates a PDF from HTML template
func TryGeneratePdf(c *gin.Context) {
	// Read HTML template
	htmlContent, err := ioutil.ReadFile("./web/html/email_template.html")
	if err != nil {
		middleware.LogError(err, "Failed to open HTML template")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to open HTML template file",
		})
		return
	}

	// Prepare template data
	year := time.Now().Year()
	templateData := map[string]string{
		"{{nama}}":       "Test User",
		"{{Opening_text}}": "Welcome to our system",
		"{{keterangan}}": "This is a test PDF generation",
		"{{Year}}":       strconv.Itoa(year),
		"{{Link}}":       "http://localhost:8080",
		"{{Nama Sistem}}": "Backend Framework",
	}

	// Replace template variables
	templateString := string(htmlContent)
	for key, value := range templateData {
		templateString = strings.Replace(templateString, key, value, -1)
	}

	// Generate PDF
	outputPath := "./temp/output.pdf"
	status := thirdparty.GeneratePdf(templateString, outputPath)
	if !status {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to generate PDF file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "PDF generated successfully",
		"data": gin.H{
			"file_path": outputPath,
		},
	})
}

// SendMail sends email using HTML template
func SendMail(c *gin.Context) {
	// Prepare recipient data
	recipientData := []thirdparty.RecipientStruct{
		{
			Name:  "Test User",
			Email: "test@gmail.com",
		},
	}

	// Prepare mail subject
	year := time.Now().Year()
	mailSubject := fmt.Sprintf("Test Mail %d", year)

	// Read HTML template
	htmlContent, err := ioutil.ReadFile("./web/html/email_template.html")
	if err != nil {
		middleware.LogError(err, "Failed to open HTML template")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to open HTML template file",
		})
		return
	}

	// Prepare template data
	templateData := map[string]string{
		"{{nama}}":       "Test User",
		"{{Opening_text}}": "Thank you for using our service",
		"{{keterangan}}": "This is a test email",
		"{{Year}}":       strconv.Itoa(year),
		"{{Link}}":       "http://localhost:8080",
		"{{Nama Sistem}}": "Backend Framework",
	}

	// Replace template variables
	templateString := string(htmlContent)
	for key, value := range templateData {
		templateString = strings.Replace(templateString, key, value, -1)
	}

	// Send email
	status := thirdparty.SendEmail(templateString, mailSubject, recipientData)
	if !status {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to send email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Email sent successfully",
		"data": gin.H{
			"recipients": len(recipientData),
			"subject":    mailSubject,
		},
	})
}

// GenerateExcel generates Excel file from user data
func GenerateExcel(c *gin.Context) {
	// Get all users from service
	users, err := service.GetAllUsers()
	if err != nil {
		middleware.LogError(err, "Failed to get users data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to retrieve users data",
		})
		return
	}

	// Check if users exist
	if len(users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "No users data available",
		})
		return
	}

	// Define Excel headers
	headers := []thirdparty.Header{
		{Text: "User", Width: 15},
		{Text: "Email", Width: 30},
		{Text: "Usergroup", Width: 20},
		{Text: "Status", Width: 10},
	}

	var excelData []map[string]interface{}
	for _, user := range users {
		status := "Inactive"
		if user.IsAktif == "Y" {
			status = "Active"
		}

		// UBAH: Menggunakan GroupID dari user dan mengkonversinya ke nama grup
		groupName := getGroupNameFromID(user.GroupID)

		excelData = append(excelData, map[string]interface{}{
			"User":      user.Username,
			"Email":     user.Email,
			"Usergroup": groupName, // UBAH: Menggunakan groupName hasil konversi
			"Status":    status,
		})
	}

	// Generate timestamp for unique filename
	timestamp := time.Now().Format("20060102_150405")
	sheetName := "User List"
	savePath := fmt.Sprintf("temp/users_%s.xlsx", timestamp)

	// Generate Excel file
	status := thirdparty.GenerateExcelFile(headers, excelData, sheetName, savePath)
	if !status {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to generate Excel file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Excel file generated successfully",
		"data": gin.H{
			"file_path":  savePath,
			"sheet_name": sheetName,
			"total_rows": len(excelData),
		},
	})
}

// ReadExcel reads and parses Excel file
func ReadExcel(c *gin.Context) {
	// Get sheet name from form
	sheetName := c.PostForm("sheetName")
	if sheetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Sheet name is required",
		})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "File is required",
		})
		return
	}

	// File validation constants
	const maxFileSize = 5 * 1024 * 1024 // 5 MB
	var allowedExtensions = []string{".xlsx", ".xls"}
	const uploadDir = "./temp/"

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// Validate file
	isValid, errMsg := middleware.ValidateFile(maxFileSize, file.Size, ext, allowedExtensions)
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "File validation failed",
			"error":   errMsg,
		})
		return
	}

	// Create temp directory if not exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		middleware.LogError(err, "Failed to create temp directory")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to create temp directory",
		})
		return
	}

	// Generate unique filename
	timestamp := time.Now().UnixNano()
	localFileName := fmt.Sprintf("upload_%d%s", timestamp, ext)
	localFilePath := filepath.Join(uploadDir, localFileName)

	// Save uploaded file
	if err := c.SaveUploadedFile(file, localFilePath); err != nil {
		middleware.LogError(err, "Failed to save uploaded file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to save uploaded file",
		})
		return
	}

	// Ensure file cleanup
	defer func() {
		if err := os.Remove(localFilePath); err != nil {
			middleware.LogError(err, "Failed to clean up temporary file")
		}
	}()

	// Read Excel file
	excelHeader, excelData, status := thirdparty.ReadExcelFile(sheetName, localFilePath)
	if !status {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to read Excel file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Excel file read successfully",
		"data": gin.H{
			"sheet_name": sheetName,
			"headers":    excelHeader,
			"rows":       excelData,
			"total_rows": len(excelData),
		},
	})
}




// PingMongo tests MongoDB connection
func PingMongo(c *gin.Context) {
	service.TestPing()
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "MongoDB ping executed - check logs for result",
	})
}