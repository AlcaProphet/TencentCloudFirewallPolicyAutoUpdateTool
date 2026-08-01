//go:build ignore
// 桌面端功能已搁置，此文件归档于 desktop/ 目录。原构建标签: //go:build desktop && darwin

package app

import (
	"log/slog"
	"os"
	"path/filepath"
)

// isAutoStartEnabled 检查 macOS LaunchAgent plist 是否存在
func isAutoStartEnabled() bool {
	plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.fwalizer.agent.plist")
	_, err := os.Stat(plistPath)
	return err == nil
}

// enableAutoStart 启用 macOS 开机自启
func enableAutoStart() {
	exePath, err := os.Executable()
	if err != nil {
		slog.Warn("获取可执行文件路径失败", "error", err)
		return
	}
	enableAutoStartDarwin(exePath)
}

// disableAutoStart 禁用 macOS 开机自启
func disableAutoStart() {
	disableAutoStartDarwin()
}
