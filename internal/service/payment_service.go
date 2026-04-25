package service

import (
	"BackendFramework/internal/config"
	"BackendFramework/internal/model"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
)

// ─── CREATE PAYMENT ────────────────────────────────────────────────────────────

func CreatePayment(projectID, userID uint, amount int64) (*model.Payment, string, error) {
	_, err := GetProjectByID(projectID)
	if err != nil {
		return nil, "", errors.New("project not found")
	}

	orderID := fmt.Sprintf("ORDER-%d-%d", projectID, time.Now().Unix())

	var s snap.Client
	env := midtrans.Sandbox
	if os.Getenv("MIDTRANS_ENV") == "production" {
		env = midtrans.Production
	}
	s.New(os.Getenv("MIDTRANS_SERVER_KEY"), env)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
	}

	snapResp, err := s.CreateTransaction(req)
	if err != nil {
		return nil, "", fmt.Errorf("midtrans error: %s", err.Error())
	}

	payment := &model.Payment{
		ProjectID: projectID,
		UserID:    userID,
		OrderID:   orderID,
		Amount:    amount,
		Status:    model.PaymentPending,
		SnapToken: snapResp.Token,
	}

	if err := config.DB.Create(payment).Error; err != nil {
		return nil, "", err
	}

	return payment, snapResp.RedirectURL, nil
}

// ─── GET ALL PAYMENTS ──────────────────────────────────────────────────────────

func GetAllPayments() ([]model.Payment, error) {
	var payments []model.Payment
	if err := config.DB.Order("created_at desc").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// ─── GET PAYMENTS BY USER ID ───────────────────────────────────────────────────

func GetPaymentsByUserID(userID uint) ([]model.Payment, error) {
	var payments []model.Payment
	if err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// ─── GET PAYMENTS BY PROJECT ID ────────────────────────────────────────────────

func GetPaymentsByProjectID(projectID uint) ([]model.Payment, error) {
	var payments []model.Payment
	if err := config.DB.Where("project_id = ?", projectID).Order("created_at desc").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// ─── GET PAYMENT BY ORDER ID ───────────────────────────────────────────────────

func GetPaymentByOrderID(orderID string) (*model.Payment, error) {
	var payment model.Payment
	if err := config.DB.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

// ─── UPDATE PAYMENT STATUS (manual by admin) ───────────────────────────────────

func UpdatePaymentStatus(orderID string, status model.PaymentStatus) error {
	result := config.DB.Model(&model.Payment{}).
		Where("order_id = ?", orderID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("payment not found")
	}
	return nil
}

// ─── HANDLE MIDTRANS NOTIFICATION (webhook) ────────────────────────────────────

func HandleMidtransNotification(notification map[string]interface{}) error {
	orderID, ok := notification["order_id"].(string)
	if !ok || orderID == "" {
		return errors.New("invalid order_id in notification")
	}

	transactionStatus, _ := notification["transaction_status"].(string)
	fraudStatus, _ := notification["fraud_status"].(string)
	paymentType, _ := notification["payment_type"].(string)

	var status model.PaymentStatus
	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			status = model.PaymentSuccess
		} else {
			status = model.PaymentFailed
		}
	case "settlement":
		status = model.PaymentSuccess
	case "deny", "cancel", "failure":
		status = model.PaymentFailed
	case "expire":
		status = model.PaymentExpired
	default:
		status = model.PaymentPending
	}

	result := config.DB.Model(&model.Payment{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":       status,
			"payment_type": paymentType,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("payment not found")
	}
	return nil
}