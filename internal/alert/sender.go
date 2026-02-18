package alert

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"sretool-fullstack/internal/model"
)

// EmailSender 邮件发送器
type EmailSender struct{}

// NewEmailSender 创建邮件发送器
func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

// SendEmail 发送告警邮件
func (s *EmailSender) SendEmail(subject, body string, config model.EmailConfig) error {
	if config.Host == "" || config.Username == "" || len(config.To) == 0 {
		return fmt.Errorf("email config is incomplete")
	}

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = config.From
	headers["To"] = strings.Join(config.To, ",")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=utf-8"

	// 构建邮件内容
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	err := smtp.SendMail(addr, auth, config.From, config.To, []byte(msg.String()))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Send 实现 AlertSender 接口
func (s *EmailSender) Send(ctx context.Context, alerts []Alert, cfg model.EmailConfig) error {
	if len(alerts) == 0 {
		return nil
	}

	subject := fmt.Sprintf("[%s] sretool-fullstack Alert - %d alerts", alerts[0].Level, len(alerts))
	body := formatAlerts(alerts)

	return s.SendEmail(subject, body, cfg)
}

// formatAlerts 格式化告警列表为邮件内容
func formatAlerts(alerts []Alert) string {
	var sb strings.Builder
	sb.WriteString("sretool-fullstack Alert Report\n")
	sb.WriteString(strings.Repeat("=", 50))
	sb.WriteString("\n\n")

	for i, alert := range alerts {
		sb.WriteString(fmt.Sprintf("#%d\n", i+1))
		sb.WriteString(fmt.Sprintf("Time: %s\n", alert.Timestamp.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("Level: %s\n", alert.Level))
		sb.WriteString(fmt.Sprintf("Category: %s\n", alert.Category))
		sb.WriteString(fmt.Sprintf("Host: %s\n", alert.Host))
		sb.WriteString(fmt.Sprintf("Message: %s\n", alert.Message))
		sb.WriteString(fmt.Sprintf("Value: %.2f %s\n", alert.Value, alert.Unit))
		sb.WriteString(fmt.Sprintf("Threshold: %.2f %s\n", alert.Threshold, alert.Unit))
		sb.WriteString("\n")
	}

	return sb.String()
}
