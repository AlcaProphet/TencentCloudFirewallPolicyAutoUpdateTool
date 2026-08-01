package app

import "os"

// Mode 运行模式
type Mode string

const (
	ModeEnv   Mode = "env"   // .env 文件驱动，无 WebUI
	ModeWebUI Mode = "webui" // SQLite + WebUI
)

// DetectMode 检测运行模式
func DetectMode(forced string) Mode {
	if forced == "env" || forced == "webui" {
		return Mode(forced)
	}
	// 自动检测：TARGETS 环境变量存在 → env 模式
	if os.Getenv("TARGETS") != "" {
		return ModeEnv
	}
	return ModeWebUI
}
