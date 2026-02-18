package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tisminSRETool/internal/model"
)

// WebhookSender Webhook 告警发送器
type WebhookSender struct {
	URL        string
	Method     string
	Headers    map[string]string
	Timeout    time.Duration
}

// NewWebhookSender 创建 Webhook 发送器
func NewWebhookSender(url string) *WebhookSender {
	return &WebhookSender{
		URL:     url,
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/json"},
		Timeout: 10 * time.Second,
	}
}

// Send 发送告警到 Webhook
func (s *WebhookSender) Send(ctx context.Context, alerts []Alert, cfg model.EmailConfig) error {
	if len(alerts) == 0 || s.URL == "" {
		return nil
	}

	// 构建请求体
	payload := map[string]interface{}{
		"alerts": alerts,
		"timestamp": time.Now().Format(time.RFC3339),
		"source":    "tisminSRETool",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, s.Method, s.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{Timeout: s.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

// DingTalkSender 钉钉告警发送器
type DingTalkSender struct {
	WebhookURL string
	Secret     string
	Timeout    time.Duration
}

// NewDingTalkSender 创建钉钉发送器
func NewDingTalkSender(webhookURL, secret string) *DingTalkSender {
	return &DingTalkSender{
		WebhookURL: webhookURL,
		Secret:     secret,
		Timeout:    10 * time.Second,
	}
}

// Send 发送告警到钉钉
func (s *DingTalkSender) Send(ctx context.Context, alerts []Alert, cfg model.EmailConfig) error {
	if len(alerts) == 0 || s.WebhookURL == "" {
		return nil
	}

	// 简单文本消息格式
	text := "## tisminSRETool 告警\n\n"
	for _, alert := range alerts {
		text += fmt.Sprintf("**%s** %s: %s\n", alert.Level, alert.Category, alert.Message)
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: s.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send dingtalk: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
