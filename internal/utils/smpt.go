
package utils

import (
	"os"
)

type SMTPConfig struct {
	Host       string
	Port       string
	AuthEmail   string
	AuthPass   string
	SenderName string
	SenderAddr string
}


func LoadSMTPConfig() *SMTPConfig {
    return &SMTPConfig{
        Host:       os.Getenv("CONFIG_SMTP_HOST"),
        Port:       os.Getenv("CONFIG_SMTP_PORT"),
        AuthEmail:  os.Getenv("CONFIG_SMTP_USERNAME"), 
        AuthPass:   os.Getenv("CONFIG_SMTP_PASSWORD"),
        SenderName: os.Getenv("CONFIG_SENDER_NAME_DEVELOPMENT"),
        SenderAddr: os.Getenv("CONFIG_SENDER_EMAIL"),
    }
}
