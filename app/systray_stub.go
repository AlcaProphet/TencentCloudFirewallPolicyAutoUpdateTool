//go:build !desktop

package app

// RunSystray 非桌面构建下为空操作
func RunSystray(openURL string, onSyncTrigger func()) {}
