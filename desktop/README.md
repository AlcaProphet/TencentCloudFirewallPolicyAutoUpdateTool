# 桌面端功能归档

> 本目录存放桌面端（macOS 系统托盘 / Windows 托盘）功能的源代码与构建资源。
> 该功能已暂时搁置开发，详见项目根目录的 [FutureDesktopDevelop.md](../FutureDesktopDevelop.md)。

## 文件清单

| 文件 | 原始位置 | 功能 |
|------|---------|------|
| `systray.go` | `app/systray.go` | 系统托盘图标与右键菜单（`fyne.io/systray`） |
| `autostart.go` | `app/autostart.go` | 开机自启通用逻辑（macOS LaunchAgent plist 生成） |
| `autostart_darwin.go` | `app/autostart_darwin.go` | macOS 开机自启检测/启用/禁用 |
| `autostart_windows.go` | `app/autostart_windows.go` | Windows 注册表 Run 键检测/启用/禁用 |
| `Info.plist` | `build/Info.plist` | macOS .app 包元数据模板 |

## 恢复步骤

将来重启桌面端开发时：

1. 将以上文件移回对应原始位置
2. 恢复 `main.go` 中被注释的托盘相关代码（搜索 `FutureDesktopDevelop.md`）
3. 恢复 `.github/workflows/release.yml` 中的桌面端构建步骤（目前内容已清空为占位）
4. 更新各设计文档中标记为"已搁置"的章节
