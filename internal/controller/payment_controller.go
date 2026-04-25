package controller

import (
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// ─── POST /v1/payment ──────────────────────────────────────────────────────────

func CreatePayment(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		return
	}

	var body struct {
		ProjectID uint  `json:"project_id" binding:"required"`
		Amount    int64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "error": "Fields 'project_id' and 'amount' are required"})
		return
	}

	payment, redirectURL, err := service.CreatePayment(body.ProjectID, userID, body.Amount)
	if err != nil {
		code := http.StatusInternalServerError
		if containsString(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":         http.StatusCreated,
		"message":      "Payment created successfully",
		"snap_token":   payment.SnapToken,
		"redirect_url": redirectURL,
		"order_id":     payment.OrderID,
	})
}

// ─── GET /v1/payment/my ────────────────────────────────────────────────────────

func GetMyPayments(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		return
	}

	payments, err := service.GetPaymentsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Payments retrieved successfully",
		"data":    payments,
		"total":   len(payments),
	})
}

// ─── GET /v1/payment/:orderId ──────────────────────────────────────────────────

func GetPaymentByOrder(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "error": "orderId is not valid"})
		return
	}

	payment, err := service.GetPaymentByOrderID(orderID)
	if err != nil {
		code := http.StatusInternalServerError
		if containsString(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Payment retrieved successfully",
		"data":    payment,
	})
}

// ─── GET /v1/payment (admin) ───────────────────────────────────────────────────

func GetAllPayments(c *gin.Context) {
	payments, err := service.GetAllPayments()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Payments retrieved successfully",
		"data":    payments,
		"total":   len(payments),
	})
}

// ─── GET /v1/payment/project/:projectId (admin) ────────────────────────────────

func GetPaymentsByProject(c *gin.Context) {
	projectID, err := parseUintParam(c, "projectId")
	if err != nil {
		return
	}

	payments, err := service.GetPaymentsByProjectID(projectID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Payments retrieved successfully",
		"data":    payments,
		"total":   len(payments),
	})
}

// ─── PATCH /v1/payment/:orderId/status (admin) ─────────────────────────────────

func UpdatePaymentStatus(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "error": "orderId is not valid"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "error": "Field 'status' is required"})
		return
	}

	if err := service.UpdatePaymentStatus(orderID, model.PaymentStatus(body.Status)); err != nil {
		code := http.StatusInternalServerError
		if containsString(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Payment status updated successfully",
		"status":  body.Status,
	})
}

// ─── POST /v1/payment/notification (public - Midtrans webhook) ─────────────────

func MidtransNotification(c *gin.Context) {
	var notification map[string]interface{}
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if err := service.HandleMidtransNotification(notification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// ─── Helper: ekstrak userID dari JWT context ───────────────────────────────────

func extractUserID(c *gin.Context) (uint, bool) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Unauthorized"})
		return 0, false
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Invalid user ID in token"})
		return 0, false
	}

	userIDParsed, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusUnauthorized, "error": "Invalid user ID format"})
		return 0, false
	}

	return uint(userIDParsed), true
}
