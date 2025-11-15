package util

import (
	"fmt"
	"net/smtp"
	"encore.app/config"
)

type EmailService interface {
	SendOTP(email, otp string) error
	SendWelcomeEmail(email, name string) error
	SendPasswordResetEmail(email, otp string) error
}

type emailService struct {
	smtpHost string
	smtpPort string
	from     string
}

func NewEmailService() EmailService {
	cfg := config.GetConfig()
	return &emailService{
		smtpHost: cfg.MailHog.SMTPServer,
		smtpPort: fmt.Sprintf("%d", cfg.MailHog.SMTPPort),
		from:     "noreply@rival.com",
	}
}

func (e *emailService) SendOTP(email, otp string) error {
	subject := "Your OTP Code - RIVAL"
	body := fmt.Sprintf(`
		<h2>Your OTP Code</h2>
		<p>Your verification code is: <strong>%s</strong></p>
		<p>This code will expire in 10 minutes.</p>
		<p>If you didn't request this, please ignore this email.</p>
	`, otp)
	
	return e.sendEmail(email, subject, body)
}

func (e *emailService) SendWelcomeEmail(email, name string) error {
	subject := "Welcome to RIVAL!"
	body := fmt.Sprintf(`
		<h2>Welcome to RIVAL, %s!</h2>
		<p>Your account has been successfully created.</p>
		<p>Start earning coins and enjoying discounts at your favorite restaurants!</p>
	`, name)
	
	return e.sendEmail(email, subject, body)
}

func (e *emailService) SendPasswordResetEmail(email, otp string) error {
	subject := "Password Reset - RIVAL"
	body := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>Your password reset code is: <strong>%s</strong></p>
		<p>This code will expire in 10 minutes.</p>
		<p>If you didn't request this, please ignore this email.</p>
	`, otp)
	
	return e.sendEmail(email, subject, body)
}

func (e *emailService) sendEmail(to, subject, body string) error {
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body)
	
	// MailHog doesn't require authentication
	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	return smtp.SendMail(addr, nil, e.from, []string{to}, []byte(msg))
}
