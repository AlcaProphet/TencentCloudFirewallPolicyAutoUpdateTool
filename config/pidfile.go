package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetPidFilePath 返回 pidfile 路径
func GetPidFilePath(dataDir string) string {
	return filepath.Join(dataDir, "fwalizer.pid")
}

// WritePidFile 写入 PID 文件，返回清理函数。若已有进程运行则返回错误。
func WritePidFile(path string) (cleanup func(), err error) {
	// 检查已有 pidfile
	if data, err := os.ReadFile(path); err == nil {
		pidStr := strings.TrimSpace(string(data))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			if processExists(pid) {
				return nil, fmt.Errorf("FWAlizer 已在运行 (PID: %d)，请先停止现有实例", pid)
			}
		}
	}

	// 写入当前 PID
	if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("写入 pidfile 失败: %w", err)
	}

	return func() { os.Remove(path) }, nil
}
