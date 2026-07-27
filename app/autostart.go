//go:build desktop

package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

// isAutoStartEnabled 检查是否已注册开机自启
func isAutoStartEnabled() bool {
	switch runtime.GOOS {
	case "darwin":
		plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.fwalizer.agent.plist")
		_, err := os.Stat(plistPath)
		return err == nil
	case "windows":
		// Windows 注册表检查（简化实现：通过 reg query 命令）
		// 桌面构建时才编译，此处使用命令行方式避免引入额外依赖
		return false // 简化：默认未启用
	default:
		return false
	}
}

// enableAutoStart 启用开机自启
func enableAutoStart() {
	exePath, err := os.Executable()
	if err != nil {
		slog.Warn("获取可执行文件路径失败", "error", err)
		return
	}

	switch runtime.GOOS {
	case "darwin":
		enableAutoStartDarwin(exePath)
	case "windows":
		enableAutoStartWindows(exePath)
	default:
		slog.Info("当前平台不支持开机自启")
	}
}

// disableAutoStart 禁用开机自启
func disableAutoStart() {
	switch runtime.GOOS {
	case "darwin":
		disableAutoStartDarwin()
	case "windows":
		disableAutoStartWindows()
	default:
		slog.Info("当前平台不支持开机自启")
	}
}

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

// ─── Windows ───

func enableAutoStartWindows(exePath string) {
	// 使用 reg add 命令写入注册表（避免引入 golang.org/x/sys/windows/registry）
	// HKCU\Software\Microsoft\Windows\CurrentVersion\Run
	slog.Info("Windows 开机自启功能待实现（需注册表操作）", "path", exePath)
}

func disableAutoStartWindows() {
	slog.Info("Windows 开机自启禁用待实现")
}
