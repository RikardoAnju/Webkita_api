// internal/model/userToken.go
package model

import "time"

type UserToken struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement"`
	UserID              string    `gorm:"column:user_id;index"`
	LastIPAddress       string    `gorm:"column:last_ip_address"`
	LastUserAgent       string    `gorm:"column:last_user_agent"`
	AccessToken         string    `gorm:"column:access_token"`
	RefreshToken        string    `gorm:"column:refresh_token"`
	RefreshTokenExpired time.Time `gorm:"column:refresh_token_expired"`
	LastLogin           time.Time `gorm:"column:last_login"`
	IsValidToken        string    `gorm:"column:is_valid_token"`
	IsRememberMe        string    `gorm:"column:is_remember_me"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}