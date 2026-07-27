//go:build !desktop

package app

// RunSystray 非桌面构建下为空操作
func RunSystray(openURL string, onSyncTrigger func()) {}

var quitCh = make(chan struct{})

// QuitCh 非桌面构建下永远不会关闭（无托盘退出按钮）
func QuitCh() <-chan struct{} { return quitCh }
