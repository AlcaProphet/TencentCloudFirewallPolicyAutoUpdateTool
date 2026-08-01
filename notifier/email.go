package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig SMTP 邮件配置
type EmailConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
	To   string
}

// EmailNotifier 邮件告警
type EmailNotifier struct {
	cfg EmailConfig
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	return &EmailNotifier{cfg: cfg}
}

// OnEvent 实现 Subscriber 接口
func (n *EmailNotifier) OnEvent(event Event) error {
	// 仅处理错误事件
	if event.Type != EventSyncError && event.Type != EventDNSFailed {
		return nil
	}

	subject := fmt.Sprintf("[FWAlizer] %s", event.Type)
	body := formatEventBody(event)

	return n.send(subject, body)
}

func (n *EmailNotifier) send(subject, body string) error {
	addr := n.cfg.Host + ":" + n.cfg.Port
	auth := smtp.PlainAuth("", n.cfg.User, n.cfg.Pass, n.cfg.Host)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		n.cfg.From, n.cfg.To, subject, body)

	return smtp.SendMail(addr, auth, n.cfg.From, strings.Split(n.cfg.To, ","), []byte(msg))
}

func formatEventBody(event Event) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("事件类型: %s\n", event.Type))
	sb.WriteString(fmt.Sprintf("时间: %s\n", event.Timestamp.Format("2006-01-02 15:04:05")))
	for k, v := range event.Data {
		sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	return sb.String()
}
