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
	url     string
	channel string
	client  *http.Client
}

// NewWebhookNotifier 创建 Webhook 通知器
func NewWebhookNotifier(url, channel string) *WebhookNotifier {
	if channel == "" {
		channel = "dingtalk"
	}
	return &WebhookNotifier{
		url:     url,
		channel: channel,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// OnEvent 实现 Subscriber 接口
func (n *WebhookNotifier) OnEvent(event Event) error {
	if event.Type != EventSyncError && event.Type != EventDNSFailed {
		return nil
	}

	content := fmt.Sprintf("[FWAlizer] %s\n%s", event.Type, formatEventBody(event))
	var payload map[string]any
	switch n.channel {
	case "feishu":
		payload = map[string]any{
			"msg_type": "text",
			"content":  map[string]string{"text": content},
		}
	case "slack":
		payload = map[string]any{"text": content}
	default: // dingtalk
		payload = map[string]any{
			"msgtype": "text",
			"text":    map[string]string{"content": content},
		}
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
