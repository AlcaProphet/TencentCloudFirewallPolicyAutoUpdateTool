package api

import (
	"log/slog"
	"strings"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/notifier"
)

// StoreLogWriter 将同步事件写入 SQLite 同步日志
type StoreLogWriter struct {
	Store *config.Store
}

// OnEvent 实现 notifier.Subscriber 接口
func (w *StoreLogWriter) OnEvent(event notifier.Event) error {
	log := config.SyncLog{Timestamp: event.Timestamp}
	if v, ok := event.Data["provider"].(string); ok {
		// 提取资源 ID：从 "tc_lighthouse(lhins-xxx)" 格式中取括号内部分
		if start := strings.Index(v, "("); start >= 0 {
			if end := strings.Index(v, ")"); end > start {
				log.Target = v[start+1 : end]
			} else {
				log.Target = v
			}
		} else {
			log.Target = v
		}
	}
	if v, ok := event.Data["domain"].(string); ok {
		log.Domain = v
	}
	switch event.Type {
	case notifier.EventSyncError:
		log.Result = "failed"
		if v, ok := event.Data["error"].(string); ok {
			log.Error = v
		}
	case notifier.EventSyncComplete:
		// 跳过全局完成事件（不携带 provider），仅记录逐域名事件
		if log.Target == "" {
			return nil
		}
		log.Result = "success"
	default:
		return nil
	}
	if err := w.Store.AddSyncLog(log); err != nil {
		slog.Warn("写入同步日志失败", "error", err)
	}
	return nil
}
