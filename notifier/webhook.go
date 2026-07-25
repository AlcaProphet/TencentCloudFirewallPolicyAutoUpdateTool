package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier Webhook 告警（支持钉钉/飞书/Slack 格式）
type WebhookNotifier struct {
	url    string
	client *http.Client
}

// NewWebhookNotifier 创建 Webhook 通知器
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// OnEvent 实现 Subscriber 接口
func (n *WebhookNotifier) OnEvent(event Event) error {
	if event.Type != EventSyncError && event.Type != EventDNSFailed {
		return nil
	}

	payload := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("[FWAlizer] %s\n%s", event.Type, formatEventBody(event)),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := n.client.Post(n.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("Webhook 发送失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook 返回状态码: %d", resp.StatusCode)
	}
	return nil
}
