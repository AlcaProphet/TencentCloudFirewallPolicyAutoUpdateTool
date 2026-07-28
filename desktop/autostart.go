//go:build ignore
// 桌面端功能已搁置，此文件归档于 desktop/ 目录。原构建标签: //go:build desktop

package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// ─── macOS ───

func enableAutoStartDarwin(exePath string) {
	plistDir := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents")
	if err := os.MkdirAll(plistDir, 0755); err != nil {
		slog.Warn("创建 LaunchAgents 目录失败", "error", err)
		return
	}
	plistPath := filepath.Join(plistDir, "com.fwalizer.agent.plist")
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.fwalizer.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
`, exePath)
	if err := os.WriteFile(plistPath, []byte(content), 0644); err != nil {
		slog.Warn("写入 plist 失败", "error", err)
		return
	}
	slog.Info("开机自启已启用（macOS LaunchAgent）")
}

func disableAutoStartDarwin() {
	plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.fwalizer.agent.plist")
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		slog.Warn("删除 plist 失败", "error", err)
		return
	}
	slog.Info("开机自启已禁用")
}
