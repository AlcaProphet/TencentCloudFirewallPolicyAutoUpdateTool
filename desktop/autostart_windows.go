//go:build ignore
// 桌面端功能已搁置，此文件归档于 desktop/ 目录。原构建标签: //go:build desktop && windows

package app

import (
	"log/slog"
	"os"

	"golang.org/x/sys/windows/registry"
)

const regRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const regValueName = "FWAlizer"

// isAutoStartEnabled 检查 Windows 注册表 Run 键
func isAutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, regRunKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(regValueName)
	return err == nil
}

// enableAutoStart 启用 Windows 开机自启（写入注册表 Run 键）
func enableAutoStart() {
	exePath, err := os.Executable()
	if err != nil {
		slog.Warn("获取可执行文件路径失败", "error", err)
		return
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, regRunKey, registry.SET_VALUE)
	if err != nil {
		slog.Warn("打开注册表 Run 键失败", "error", err)
		return
	}
	defer k.Close()

	if err := k.SetStringValue(regValueName, exePath); err != nil {
		slog.Warn("写入注册表 Run 键失败", "error", err)
		return
	}
	slog.Info("开机自启已启用（Windows 注册表）")
}

// disableAutoStart 禁用 Windows 开机自启（删除注册表 Run 键）
func disableAutoStart() {
	k, err := registry.OpenKey(registry.CURRENT_USER, regRunKey, registry.SET_VALUE)
	if err != nil {
		slog.Warn("打开注册表 Run 键失败", "error", err)
		return
	}
	defer k.Close()

	if err := k.DeleteValue(regValueName); err != nil {
		if err != registry.ErrNotExist {
			slog.Warn("删除注册表 Run 键失败", "error", err)
			return
		}
		// 值不存在 = 已禁用，幂等成功
	}
	slog.Info("开机自启已禁用")
}
