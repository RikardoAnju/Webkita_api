package model

import "gorm.io/gorm"

type PaymentStatus string

const (
    PaymentPending  PaymentStatus = "pending"
    PaymentSuccess  PaymentStatus = "success"
    PaymentFailed   PaymentStatus = "failed"
    PaymentExpired  PaymentStatus = "expired"
)

type Payment struct {
    gorm.Model
    ProjectID   uint          `json:"project_id"`
    UserID      uint          `json:"user_id"`
    OrderID     string        `json:"order_id" gorm:"uniqueIndex"`
    Amount      int64         `json:"amount"`
    Status      PaymentStatus `json:"status" gorm:"default:'pending'"`
    SnapToken   string        `json:"snap_token"`
    PaymentType string        `json:"payment_type"`
}