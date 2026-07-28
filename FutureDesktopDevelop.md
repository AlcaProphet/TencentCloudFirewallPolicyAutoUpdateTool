# 桌面端开发搁置记录

> 本文档记录桌面端（macOS 系统托盘 / Windows 托盘）功能已实现部分、阻塞问题及搁置决定。
> 更新时间：2026-07-28

---

## 一、已实现部分

| 模块 | 文件 | 功能 |
|------|------|------|
| 系统托盘 | `app/systray.go` | 使用 `fyne.io/systray` 实现托盘图标、右键菜单（状态/打开面板/立即同步/开机自启/退出） |
| 非桌面存根 | `app/systray_stub.go` | `//go:build !desktop` 空操作存根 |
| macOS 开机自启 | `app/autostart.go` + `app/autostart_darwin.go` | LaunchAgent plist 写入/删除 |
| Windows 开机自启 | `app/autostart.go` + `app/autostart_windows.go` | 注册表 Run 键写入/删除 |
| .app 打包 | `build/Info.plist` | macOS 应用包元数据（CFBundleIdentifier、LSUIElement 等） |
| CI 构建 | `.github/workflows/release.yml` | macOS .app 打包、codesign 签名、ditto 压缩；Windows .exe 构建 |

---

## 二、遇到的阻塞问题

### 核心错误
macOS 构建的 `.app` 解压后出现以下两种错误：
1. **"你没有兼容的应用打开此程序"** — Finder 双击时 Launch Services 报 `-1712`
2. **"The application cannot be opened for an unexpected reason"** — `open` 命令报 `RBSRequestErrorDomain Code=5`, 底层 `POSIX error 153`

### 问题分析
- **代码签名**：Go 链接器的 ad-hoc 签名不与 `.app` 内的 Info.plist 绑定，导致 `Identifier=a.out`、`Info.plist=not bound`
- **Launch Services 兼容性**：即使执行了 `codesign --force --deep --sign -` 完整 ad-hoc 签名（修复了 identifier 绑定），Launch Services 仍报 `-1712` 拒绝启动。直接运行二进制（`./FWAlizer.app/Contents/MacOS/fwalizer`）**可以正常工作**，托盘图标也能正常显示
- **根本原因**：`fyne.io/systray` 依赖 AppKit 的 `NSStatusBar`，要求进程以正常 Cocoa 应用方式启动。但 `.app` 包的 Launch Services 注册 / Gatekeeper 安全策略层面存在兼容性问题，导致 macOS 拒绝通过 Finder/open 启动该应用

### 已尝试的修复步骤

| 尝试 | 方法 | 结果 |
|------|------|------|
| 1 | `zip -r` 打包 → 解压 .app 损坏 | 失败 — zip 不保留 macOS 权限 |
| 2 | `ditto -c -k --keepParent` 替代 zip | 改善 — 权限保留，但问题依旧 |
| 3 | `codesign --force --deep --sign -` 完整签名 | 改善 — Identifier/Info.plist 绑定正确，但 Launch Services 仍拒绝 |
| 4 | 移除 `LSUIElement` 键 | 无效 — 错误不变 |
| 5 | 移除所有签名（`codesign --remove-signature`） | 更差 — 报 POSIX 153 launchd spawn 失败 |
| 6 | `lsregister -f` 重新注册 Launch Services | 无效 |
| 7 | `xattr -d com.apple.quarantine` 清除隔离 | 无效 — 非隔离问题 |

### 关键发现
直接通过命令行执行 `.app` 内的二进制（`FWAlizer.app/Contents/MacOS/fwalizer`）**完全正常**，系统托盘图标正常显示。问题**仅限于 Launch Services 启动路径**（Finder 双击 / `open` 命令）。

---

## 三、搁置决定

**决定时间**：2026-07-28

**决定内容**：暂时搁置所有多平台桌面端（macOS/Windows 系统托盘）开发计划。

**依据**：
1. 核心问题（Launch Services 兼容性）需要深入调试 `fyne.io/systray` 与 macOS 安全模型的交互，预计耗时较长
2. 项目核心价值（DNS→防火墙同步）在无桌面端的情况下完全可用（.env 模式 / Docker / WebUI）
3. 继续投入到此问题会阻塞其他更高优先级的开发工作

**后续重启思路**：
- 调查 `fyne.io/systray` 在 macOS 15+ 上的已知问题
- 考虑替换为纯命令行方式（无托盘），或改用其他托盘库
- 评估是否需要完整的 Apple Developer ID 签名 + 公证流程才能解决 Launch Services 兼容性
- 参考其他 Go 托盘应用（如 syncthing、rclone 等）的 macOS 打包方案

---

## 四、代码归档位置

桌面端相关代码已从活跃代码中移除，归档至 `desktop/` 目录：

```
desktop/
├── README.md                 # 归档说明
├── Info.plist                # macOS .app 元数据模板（从 build/ 移入）
├── systray.go                # 系统托盘实现（从 app/ 移入）
├── autostart_darwin.go       # macOS 开机自启（从 app/ 移入）
└── autostart_windows.go      # Windows 开机自启（从 app/ 移入）
```

保留的文件（无需归档，因其为非桌面构建的正常存根）：
- `app/systray_stub.go` — `!desktop` 构建下的空操作存根，WebUI/.env 模式依赖

---

## 五、受影响的设计文档

以下文档的相关章节已标记为"已搁置"并指向本文档：
- `AGENTS.md` — 十一、代码规范（桌面端构建标签说明）
- `Design1.md` — 四/五章（运行模式/多端部署）、九章（Phase 4）、十章（技术决策）
- `Build1.md` — Step 16（桌面端）
- `Build2.md` — Step 11（FEA-06 systray）
- `README.md` — 桌面端系统托盘章节
