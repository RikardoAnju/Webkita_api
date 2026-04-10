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
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#f0f2f5;padding:40px 16px;">
    <tr><td align="center">
      <table width="560" cellpadding="0" cellspacing="0"
             style="background:#fff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">

        <!-- Header -->
        <tr>
          <td style="padding:28px 48px;border-bottom:1px solid #f3f4f6;">
            <table width="100%"><tr>
              <td><span style="font-size:20px;font-weight:700;color:#111827;letter-spacing:-0.3px;">{{.AppName}}</span></td>
              <td align="right">
                <span style="background:#EFF6FF;color:#1d4ed8;font-size:12px;font-weight:600;padding:5px 14px;border-radius:20px;border:1px solid #bfdbfe;">Platform Teknologi #1</span>
              </td>
            </tr></table>
          </td>
        </tr>

        <!-- Accent bar -->
        <tr><td style="height:4px;background:linear-gradient(90deg,#111827 0%,#2563EB 60%,#93c5fd 100%);"></td></tr>

        <!-- Body -->
        <tr>
          <td style="padding:40px 48px;">

            <p style="margin:0 0 6px;font-size:13px;color:#6b7280;font-weight:500;text-transform:uppercase;letter-spacing:0.8px;">Verifikasi Akun</p>
            <h2 style="margin:0 0 16px;font-size:24px;font-weight:700;color:#111827;line-height:1.3;">Halo, <span style="color:#2563EB;">{{.Username}}</span>!</h2>
            <p style="margin:0 0 28px;color:#4b5563;line-height:1.7;font-size:15px;">
              Terima kasih telah bergabung di <strong style="color:#111827;">{{.AppName}}</strong> — platform marketplace yang menghubungkan Anda dengan developer profesional terbaik. Satu langkah lagi untuk mengaktifkan akun Anda.
            </p>

            <!-- Stats strip -->
            

            <!-- Info box -->
            <table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
              <tr>
                <td style="background:#EFF6FF;border-left:3px solid #2563EB;border-radius:0 8px 8px 0;padding:14px 18px;">
                  <p style="margin:0;color:#1d4ed8;font-size:13px;line-height:1.6;">
                    Link verifikasi ini berlaku selama <strong>24 jam</strong> sejak email dikirim. Jika sudah kedaluwarsa, Anda dapat meminta link baru melalui halaman login.
                  </p>
                </td>
              </tr>
            </table>

            <!-- CTA Buttons -->
            <table width="100%" cellpadding="0" cellspacing="0">
              <tr>
                <td style="padding-right:12px;" width="50%">
                  <a href="{{.VerifyURL}}"
                     style="display:block;text-align:center;padding:14px 20px;background:#111827;color:#fff;text-decoration:none;border-radius:8px;font-size:14px;font-weight:600;">
                    Verifikasi Email Saya &rarr;
                  </a>
                </td>
            
              </tr>
            </table>

          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="border-top:1px solid #f3f4f6;padding:24px 48px;background:#f9fafb;">
            <table width="100%"><tr>
              <td>
                <p style="margin:0 0 2px;font-weight:700;font-size:13px;color:#111827;">{{.AppName}}</p>
                <p style="margin:0;color:#9ca3af;font-size:11px;">Hubungkan Bisnis Anda dengan Developer Terbaik</p>
              </td>
              <td align="right">
                <p style="margin:0;color:#9ca3af;font-size:11px;text-align:right;">
                  Beranda &nbsp;&middot;&nbsp; Cara Kerja &nbsp;&middot;&nbsp; Harga<br>
                  <a href="{{.AppURL}}" style="color:#2563EB;text-decoration:none;">{{.AppURL}}</a>
                </p>
              </td>
            </tr></table>
            <p style="margin:16px 0 0;color:#d1d5db;font-size:11px;text-align:center;">
              Jika Anda tidak mendaftar di {{.AppName}}, abaikan email ini. &nbsp;&middot;&nbsp; &copy; {{.Year}} {{.AppName}}
            </p>
          </td>
        </tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`
