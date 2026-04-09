package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type EmailService struct {
	apiKey     string
	fromEmail  string
	fromName   string
	appName    string
	appURL     string
	httpClient *http.Client
}

func NewEmailService() *EmailService {
	return &EmailService{
		apiKey:     os.Getenv("MAILERSEND_API_KEY"),
		fromEmail:  os.Getenv("MAIL_FROM_EMAIL"),
		fromName:   os.Getenv("MAIL_FROM_NAME"),
		appName:    getEnv("APP_NAME", "WebKita"),
		appURL:     os.Getenv("APP_URL"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ─── MailerSend Payload ───────────────────────────────────────────────────────

type mailerSendPayload struct {
	From    mailerContact   `json:"from"`
	To      []mailerContact `json:"to"`
	Subject string          `json:"subject"`
	HTML    string          `json:"html"`
	Text    string          `json:"text"`
}

type mailerContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// ─── Public Methods ───────────────────────────────────────────────────────────

// SendVerificationEmail mengirim email verifikasi ke user yang baru daftar
func (s *EmailService) SendVerificationEmail(toEmail, username, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s&email=%s", s.appURL, token, toEmail)

	html, err := s.renderHTML(verificationEmailTemplate, map[string]string{
		"AppName":   s.appName,
		"Username":  username,
		"VerifyURL": verifyURL,
		"AppURL":    s.appURL,
		"Year":      fmt.Sprintf("%d", time.Now().Year()),
	})
	if err != nil {
		return fmt.Errorf("render template gagal: %w", err)
	}

	plain := fmt.Sprintf(
		"Halo %s,\n\nVerifikasi email Anda:\n%s\n\nLink berlaku 24 jam.\n\n%s",
		username, verifyURL, s.appName,
	)

	return s.send(toEmail, username, fmt.Sprintf("[%s] Verifikasi Email Anda", s.appName), html, plain)
}

// SendResendVerificationEmail - kirim ulang verifikasi
func (s *EmailService) SendResendVerificationEmail(toEmail, username, token string) error {
	return s.SendVerificationEmail(toEmail, username, token)
}

// ─── Internal ─────────────────────────────────────────────────────────────────

func (s *EmailService) send(toEmail, toName, subject, html, text string) error {
	payload := mailerSendPayload{
		From:    mailerContact{Email: s.fromEmail, Name: s.fromName},
		To:      []mailerContact{{Email: toEmail, Name: toName}},
		Subject: subject,
		HTML:    html,
		Text:    text,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.mailersend.com/v1/email", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gagal kirim ke MailerSend: %w", err)
	}
	defer resp.Body.Close()

	// MailerSend mengembalikan 202 Accepted jika berhasil
	if resp.StatusCode != http.StatusAccepted {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("❌ MailerSend error [HTTP %d]: %v", resp.StatusCode, errResp)
		return fmt.Errorf("mailersend error, status: %d", resp.StatusCode)
	}

	log.Printf("✅ Email terkirim ke: %s", toEmail)
	return nil
}

func (s *EmailService) renderHTML(tmplStr string, data map[string]string) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ─── Email Template ───────────────────────────────────────────────────────────

const verificationEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
  <title>Verifikasi Email - {{.AppName}}</title>
</head>
<body style="margin:0;padding:0;background:#f0f2f5;font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#f0f2f5;padding:48px 16px;">
    <tr><td align="center">
      <table width="560" cellpadding="0" cellspacing="0"
             style="background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 20px rgba(0,0,0,0.08);">

        <!-- Header -->
        <tr>
          <td style="background:linear-gradient(135deg,#4F46E5 0%,#7C3AED 100%);padding:36px 48px;text-align:center;">
            <h1 style="margin:0;color:#fff;font-size:26px;font-weight:700;">{{.AppName}}</h1>
            <p style="margin:8px 0 0;color:rgba(255,255,255,0.75);font-size:14px;">Verifikasi Akun Anda</p>
          </td>
        </tr>

        <!-- Body -->
        <tr>
          <td style="padding:40px 48px;">
            <h2 style="margin:0 0 16px;color:#1e1e2e;font-size:20px;">Halo, {{.Username}}! 👋</h2>
            <p style="margin:0 0 24px;color:#52525b;line-height:1.7;font-size:15px;">
              Terima kasih telah mendaftar di <strong style="color:#4F46E5;">{{.AppName}}</strong>.
              Satu langkah lagi! Verifikasi email Anda untuk mengaktifkan akun.
            </p>

            <!-- Info -->
            <table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
              <tr>
                <td style="background:#F5F3FF;border-left:4px solid #4F46E5;border-radius:0 6px 6px 0;padding:14px 18px;">
                  <p style="margin:0;color:#6D28D9;font-size:13px;">
                    ⏰ Link verifikasi berlaku <strong>24 jam</strong> sejak email ini dikirim.
                  </p>
                </td>
              </tr>
            </table>

            <!-- Button -->
            <table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:32px;">
              <tr>
                <td align="center">
                  <a href="{{.VerifyURL}}"
                     style="display:inline-block;padding:15px 44px;background:linear-gradient(135deg,#4F46E5,#7C3AED);color:#fff;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;">
                    ✅ Verifikasi Email Saya
                  </a>
                </td>
              </tr>
            </table>

            <!-- Fallback -->
            <table width="100%" cellpadding="0" cellspacing="0">
              <tr>
                <td style="background:#f9fafb;border-radius:6px;padding:16px;border:1px solid #e5e7eb;">
                  <p style="margin:0 0 6px;color:#9ca3af;font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:0.5px;">
                    Atau salin link ini:
                  </p>
                  <a href="{{.VerifyURL}}" style="color:#4F46E5;font-size:12px;word-break:break-all;text-decoration:none;">
                    {{.VerifyURL}}
                  </a>
                </td>
              </tr>
            </table>
          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="border-top:1px solid #f3f4f6;padding:24px 48px;text-align:center;">
            <p style="margin:0 0 4px;color:#9ca3af;font-size:12px;">
              Jika Anda tidak mendaftar di {{.AppName}}, abaikan email ini.
            </p>
            <p style="margin:0;color:#d1d5db;font-size:11px;">
              © {{.Year}} {{.AppName}} ·
              <a href="{{.AppURL}}" style="color:#4F46E5;text-decoration:none;">{{.AppURL}}</a>
            </p>
          </td>
        </tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`