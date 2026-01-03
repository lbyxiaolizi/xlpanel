package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"xlpanel/internal/core"
)

type EmailService struct {
	config core.AppConfig
}

func NewEmailService(config core.AppConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// SendEmail sends an email using SMTP
func (s *EmailService) SendEmail(to, subject, htmlBody string) error {
	if s.config.SMTPHost == "" {
		return fmt.Errorf("SMTP not configured")
	}

	from := s.config.SMTPFrom
	if from == "" {
		from = "noreply@xlpanel.local"
	}

	// Create message
	msg := s.buildEmailMessage(from, to, subject, htmlBody)

	// Setup authentication
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func (s *EmailService) buildEmailMessage(from, to, subject, htmlBody string) string {
	var msg bytes.Buffer

	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.SMTPFromName, from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	return msg.String()
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *EmailService) SendWelcomeEmail(to, name, lang string) error {
	subject := "Welcome to XLPanel"
	if lang == "zh" {
		subject = "欢迎使用 XLPanel"
	}

	tmpl := s.getWelcomeTemplate(lang)
	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Name": name}); err != nil {
		return err
	}

	return s.SendEmail(to, subject, body.String())
}

// Send2FAEmail sends a 2FA setup email
func (s *EmailService) Send2FAEmail(to, name, secret, lang string) error {
	subject := "Two-Factor Authentication Setup"
	if lang == "zh" {
		subject = "双因素认证设置"
	}

	tmpl := s.get2FATemplate(lang)
	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{
		"Name":   name,
		"Secret": secret,
	}); err != nil {
		return err
	}

	return s.SendEmail(to, subject, body.String())
}

// SendInvoiceEmail sends an invoice notification email
func (s *EmailService) SendInvoiceEmail(to, name, invoiceID, amount, currency, lang string) error {
	subject := fmt.Sprintf("Invoice %s", invoiceID)
	if lang == "zh" {
		subject = fmt.Sprintf("账单 %s", invoiceID)
	}

	tmpl := s.getInvoiceTemplate(lang)
	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{
		"Name":      name,
		"InvoiceID": invoiceID,
		"Amount":    amount,
		"Currency":  currency,
	}); err != nil {
		return err
	}

	return s.SendEmail(to, subject, body.String())
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to, name, resetToken, lang string) error {
	subject := "Password Reset Request"
	if lang == "zh" {
		subject = "密码重置请求"
	}

	tmpl := s.getPasswordResetTemplate(lang)
	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{
		"Name":       name,
		"ResetToken": resetToken,
	}); err != nil {
		return err
	}

	return s.SendEmail(to, subject, body.String())
}

func (s *EmailService) getWelcomeTemplate(lang string) *template.Template {
	enTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Welcome to XLPanel!</h1>
		</div>
		<div class="content">
			<p>Hello {{.Name}},</p>
			<p>Thank you for registering with XLPanel. Your account has been successfully created.</p>
			<p>You can now log in and start managing your services.</p>
			<p>Best regards,<br>The XLPanel Team</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`

	zhTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>欢迎使用 XLPanel！</h1>
		</div>
		<div class="content">
			<p>您好 {{.Name}}，</p>
			<p>感谢您注册 XLPanel。您的账户已成功创建。</p>
			<p>您现在可以登录并开始管理您的服务。</p>
			<p>此致<br>XLPanel 团队</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. 保留所有权利。</p>
		</div>
	</div>
</body>
</html>
`

	if lang == "zh" {
		return template.Must(template.New("welcome").Parse(zhTemplate))
	}
	return template.Must(template.New("welcome").Parse(enTemplate))
}

func (s *EmailService) get2FATemplate(lang string) *template.Template {
	enTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.secret { background: #fff; padding: 15px; border: 2px dashed #2563eb; margin: 20px 0; font-family: monospace; font-size: 16px; text-align: center; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Two-Factor Authentication Setup</h1>
		</div>
		<div class="content">
			<p>Hello {{.Name}},</p>
			<p>You have enabled two-factor authentication for your account. Please save your secret key securely:</p>
			<div class="secret">{{.Secret}}</div>
			<p>Use this secret in your authenticator app (such as Google Authenticator or Authy).</p>
			<p>Best regards,<br>The XLPanel Team</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`

	zhTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.secret { background: #fff; padding: 15px; border: 2px dashed #2563eb; margin: 20px 0; font-family: monospace; font-size: 16px; text-align: center; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>双因素认证设置</h1>
		</div>
		<div class="content">
			<p>您好 {{.Name}}，</p>
			<p>您已为账户启用双因素认证。请妥善保存您的密钥：</p>
			<div class="secret">{{.Secret}}</div>
			<p>请在身份验证应用（如 Google Authenticator 或 Authy）中使用此密钥。</p>
			<p>此致<br>XLPanel 团队</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. 保留所有权利。</p>
		</div>
	</div>
</body>
</html>
`

	if lang == "zh" {
		return template.Must(template.New("2fa").Parse(zhTemplate))
	}
	return template.Must(template.New("2fa").Parse(enTemplate))
}

func (s *EmailService) getInvoiceTemplate(lang string) *template.Template {
	enTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.invoice-details { background: #fff; padding: 15px; margin: 20px 0; border-left: 4px solid #2563eb; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Invoice Notification</h1>
		</div>
		<div class="content">
			<p>Hello {{.Name}},</p>
			<p>A new invoice has been generated for your account:</p>
			<div class="invoice-details">
				<p><strong>Invoice ID:</strong> {{.InvoiceID}}</p>
				<p><strong>Amount:</strong> {{.Amount}} {{.Currency}}</p>
			</div>
			<p>Please log in to your account to view and pay the invoice.</p>
			<p>Best regards,<br>The XLPanel Team</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`

	zhTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.invoice-details { background: #fff; padding: 15px; margin: 20px 0; border-left: 4px solid #2563eb; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>账单通知</h1>
		</div>
		<div class="content">
			<p>您好 {{.Name}}，</p>
			<p>您的账户已生成新账单：</p>
			<div class="invoice-details">
				<p><strong>账单编号：</strong> {{.InvoiceID}}</p>
				<p><strong>金额：</strong> {{.Amount}} {{.Currency}}</p>
			</div>
			<p>请登录您的账户以查看和支付账单。</p>
			<p>此致<br>XLPanel 团队</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. 保留所有权利。</p>
		</div>
	</div>
</body>
</html>
`

	if lang == "zh" {
		return template.Must(template.New("invoice").Parse(zhTemplate))
	}
	return template.Must(template.New("invoice").Parse(enTemplate))
}

func (s *EmailService) getPasswordResetTemplate(lang string) *template.Template {
	enTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.token { background: #fff; padding: 15px; border: 2px dashed #dc2626; margin: 20px 0; font-family: monospace; font-size: 16px; text-align: center; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Password Reset</h1>
		</div>
		<div class="content">
			<p>Hello {{.Name}},</p>
			<p>We received a request to reset your password. Use the following token to reset it:</p>
			<div class="token">{{.ResetToken}}</div>
			<p>If you did not request a password reset, please ignore this email.</p>
			<p>Best regards,<br>The XLPanel Team</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`

	zhTemplate := `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(to right, #2563eb, #1e40af); color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9fafb; }
		.token { background: #fff; padding: 15px; border: 2px dashed #dc2626; margin: 20px 0; font-family: monospace; font-size: 16px; text-align: center; }
		.footer { text-align: center; padding: 20px; color: #6b7280; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>密码重置</h1>
		</div>
		<div class="content">
			<p>您好 {{.Name}}，</p>
			<p>我们收到了重置密码的请求。请使用以下令牌重置密码：</p>
			<div class="token">{{.ResetToken}}</div>
			<p>如果您没有请求重置密码，请忽略此邮件。</p>
			<p>此致<br>XLPanel 团队</p>
		</div>
		<div class="footer">
			<p>&copy; 2024 XLPanel. 保留所有权利。</p>
		</div>
	</div>
</body>
</html>
`

	if lang == "zh" {
		return template.Must(template.New("password_reset").Parse(zhTemplate))
	}
	return template.Must(template.New("password_reset").Parse(enTemplate))
}
