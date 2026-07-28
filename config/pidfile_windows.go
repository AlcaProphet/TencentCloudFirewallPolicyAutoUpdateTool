//go:build windows

package config

import (
	"golang.org/x/sys/windows"
)

// stillActiveExitCode 是 GetExitCodeProcess 对仍在运行的进程返回的状态码
const stillActiveExitCode = 259

func processExists(pid int) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var code uint32
	err = windows.GetExitCodeProcess(h, &code)
	return err == nil && code == uint32(stillActiveExitCode)
}
