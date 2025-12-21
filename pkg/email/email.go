package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"strings"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	UseTLS   bool // Use STARTTLS (port 587) or Implicit TLS (port 465)
}

// Sender defines the email sending interface
type Sender interface {
	Send(to, subject, body string) error
	SendHTML(to, subject, htmlBody string) error
	SendTemplate(to, subject, templateName string, data interface{}) error
}

// SMTPSender implements email sending via SMTP
type SMTPSender struct {
	config    Config
	templates map[string]*template.Template
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(cfg Config) *SMTPSender {
	sender := &SMTPSender{
		config:    cfg,
		templates: make(map[string]*template.Template),
	}

	// Register default templates
	sender.registerDefaultTemplates()

	return sender
}

// registerDefaultTemplates registers built-in email templates
func (s *SMTPSender) registerDefaultTemplates() {
	// Email verification template
	s.templates["verify_email"] = template.Must(template.New("verify_email").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
        .code { background: #e5e7eb; padding: 10px 20px; font-size: 24px; letter-spacing: 4px; border-radius: 4px; display: inline-block; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to GoMarketplace!</h1>
        </div>
        <div class="content">
            <h2>Verify Your Email Address</h2>
            <p>Hi {{.Username}},</p>
            <p>Thanks for signing up! Please verify your email address by clicking the button below:</p>
            <p style="text-align: center;">
                <a href="{{.VerifyURL}}" class="button">Verify Email</a>
            </p>
            <p>Or use this verification code:</p>
            <p style="text-align: center;">
                <span class="code">{{.Code}}</span>
            </p>
            <p>This link will expire in {{.ExpiresIn}}.</p>
            <p>If you didn't create an account, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; GoMarketplace. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`))

	// Password reset template
	s.templates["reset_password"] = template.Must(template.New("reset_password").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #DC2626; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #DC2626; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
        .warning { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 10px 15px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <h2>Reset Your Password</h2>
            <p>Hi {{.Username}},</p>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <p style="text-align: center;">
                <a href="{{.ResetURL}}" class="button">Reset Password</a>
            </p>
            <p>This link will expire in {{.ExpiresIn}}.</p>
            <div class="warning">
                <strong>Security Notice:</strong> If you didn't request a password reset, please ignore this email or contact support if you're concerned about your account security.
            </div>
        </div>
        <div class="footer">
            <p>&copy; GoMarketplace. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`))
}

// Send sends a plain text email
func (s *SMTPSender) Send(to, subject, body string) error {
	return s.sendMail(to, subject, body, "text/plain")
}

// SendHTML sends an HTML email
func (s *SMTPSender) SendHTML(to, subject, htmlBody string) error {
	return s.sendMail(to, subject, htmlBody, "text/html")
}

// SendTemplate sends an email using a registered template
func (s *SMTPSender) SendTemplate(to, subject, templateName string, data interface{}) error {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendHTML(to, subject, buf.String())
}

// sendMail sends an email with the specified content type
func (s *SMTPSender) sendMail(to, subject, body, contentType string) error {
	from := s.config.From
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=UTF-8", contentType)

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Port 465 uses implicit TLS (direct TLS connection)
	// Port 587 uses STARTTLS (connect plain, then upgrade to TLS)
	if s.config.Port == 465 {
		return s.sendMailImplicitTLS(addr, []byte(msg.String()), to)
	}

	// Port 587 or other: use STARTTLS
	return s.sendMailSTARTTLS(addr, []byte(msg.String()), to)
}

// sendMailSTARTTLS sends email using STARTTLS (port 587)
// This connects in plain text first, then upgrades to TLS
func (s *SMTPSender) sendMailSTARTTLS(addr string, msg []byte, to string) error {
	// Connect in plain text
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Say hello
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("HELO failed: %w", err)
	}

	// Check if STARTTLS is supported
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// Authenticate if credentials provided
	if s.config.Username != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// sendMailImplicitTLS sends email using implicit TLS (port 465)
// This establishes TLS from the start
func (s *SMTPSender) sendMailImplicitTLS(addr string, msg []byte, to string) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate if credentials provided
	if s.config.Username != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// VerifyEmailData holds data for email verification template
type VerifyEmailData struct {
	Username  string
	VerifyURL string
	Code      string
	ExpiresIn string
}

// ResetPasswordData holds data for password reset template
type ResetPasswordData struct {
	Username  string
	ResetURL  string
	ExpiresIn string
}
