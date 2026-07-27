//go:build desktop

package app

import (
	"log/slog"
	"os/exec"
	"runtime"

	"fyne.io/systray"
)

var quitCh = make(chan struct{})

// QuitCh 返回退出信号 channel（供 main.go 监听）
func QuitCh() <-chan struct{} { return quitCh }

// RunSystray 启动系统托盘（仅桌面端编译）
func RunSystray(openURL string, onSyncTrigger func()) {
	systray.Run(func() {
		onSystrayReady(openURL, onSyncTrigger)
	}, func() {
		onSystrayExit()
	})
}

func onSystrayReady(openURL string, onSyncTrigger func()) {
	systray.SetTitle("FWAlizer")
	systray.SetTooltip("FWAlizer - 防火墙 DNS 同步工具")

	// 菜单项
	mStatus := systray.AddMenuItem("● 运行中", "服务状态")
	mStatus.Disable()
	systray.AddSeparator()

	mOpen := systray.AddMenuItem("打开配置面板", "在浏览器中打开 WebUI")
	mSync := systray.AddMenuItem("立即同步", "手动触发一次同步")
	systray.AddSeparator()
	mAutoStart := systray.AddMenuItemCheckbox("开机自启", "启动时自动运行", isAutoStartEnabled())
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出 FWAlizer")

	// 自动打开浏览器
	openBrowser(openURL)

	// 处理菜单点击
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser(openURL)
			case <-mSync.ClickedCh:
				slog.Info("手动触发同步")
				onSyncTrigger()
			case <-mAutoStart.ClickedCh:
				if mAutoStart.Checked() {
					disableAutoStart()
					mAutoStart.Uncheck()
				} else {
					enableAutoStart()
					mAutoStart.Check()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				close(quitCh)
				return
			}
		}
	}()
}

func onSystrayExit() {
	slog.Info("系统托盘退出")
}

// openBrowser 打开默认浏览器
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // linux
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		slog.Warn("打开浏览器失败", "error", err)
	}
}
