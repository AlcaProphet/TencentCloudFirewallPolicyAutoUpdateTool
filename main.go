package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alcaprophet/fwalizer/app"
	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/dns"
	"github.com/alcaprophet/fwalizer/notifier"
	"github.com/alcaprophet/fwalizer/provider"
	"github.com/alcaprophet/fwalizer/syncer"
	"github.com/alcaprophet/fwalizer/webui"
	webapi "github.com/alcaprophet/fwalizer/webui/api"
)

func main() {
	// CLI 子命令优先
	if app.RunCLI(os.Args) {
		return
	}

	// 检测模式
	mode := app.DetectMode(os.Getenv("FWALIZER_MODE"))

	var cfg *config.Config

	switch mode {
	case app.ModeEnv:
		// 从 .env 加载
		var err error
		cfg, err = config.LoadEnv(".env")
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载 .env 失败: %v\n", err)
			os.Exit(1)
		}
	case app.ModeWebUI:
		// WebUI 模式：SQLite + HTTP Server + Syncer
		dataDir := getDataDir()
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建数据目录失败: %v\n", err)
			os.Exit(1)
		}
		dbPath := filepath.Join(dataDir, "config.db")
		store, err := config.OpenStore(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		cfg, err = store.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		// 初始化日志（同时输出到 stdout 和 WebUI 日志流）
		logBroadcaster := webapi.NewLogBroadcaster()
		app.InitLoggerWithBroadcaster(cfg.LogLevel, logBroadcaster)

		// 启动 WebUI 服务器
		srv := webui.NewServer(store, cfg.WebUIPort)
		srv.SetLogBroadcaster(logBroadcaster)

		// 创建同步引擎（初始 Provider 可为空，等待用户通过 WebUI 配置后热重载生效）
		provider.SetCredentials(cfg.TCAccessID, cfg.TCAccessKey, cfg.AliAccessID, cfg.AliAccessKey)
		pool := provider.NewClientPool()
		var providers []provider.Provider
		for _, t := range cfg.Targets {
			p, err := provider.NewProvider(t, t.ID, pool)
			if err != nil {
				fmt.Fprintf(os.Stderr, "创建 Provider 失败: %v\n", err)
				os.Exit(1)
			}
			providers = append(providers, p)
		}
		resolver := dns.NewResolver(cfg.DNS, cfg.DNSTimeout)
		s := syncer.New(cfg, providers, resolver)

		// 将 Syncer 和 EventBus 传入 WebUI（支持 status/trigger/dryrun/SSE）
		srv.SetSyncer(s, s.EventBus())

		// 同步日志写入：订阅 sync:complete 和 sync:error 事件
		logWriter := &webapi.StoreLogWriter{Store: store}
		s.EventBus().Subscribe(notifier.EventSyncComplete, logWriter)
		s.EventBus().Subscribe(notifier.EventSyncError, logWriter)

		// 接通热重载：WebUI 修改配置后重新加载并通知 Syncer
		srv.SetReloadFunc(func() {
			newCfg, err := store.LoadConfig()
			if err != nil {
				slog.Error("重载配置失败", "error", err)
				return
			}
			// 更新凭据
			provider.SetCredentials(newCfg.TCAccessID, newCfg.TCAccessKey, newCfg.AliAccessID, newCfg.AliAccessKey)
			// 重建 ClientPool 和 Provider 列表
			newPool := provider.NewClientPool()
			var newProviders []provider.Provider
			for _, t := range newCfg.Targets {
				p, err := provider.NewProvider(t, t.ID, newPool)
				if err != nil {
					slog.Error("重建 Provider 失败", "target", t.ResourceID, "error", err)
					continue
				}
				newProviders = append(newProviders, p)
			}
			s.ReloadProviders(newProviders)
			s.Reload(newCfg)
		})

		if len(providers) == 0 {
			slog.Info("WebUI 已启动，请通过浏览器配置云资源凭据和目标", "port", cfg.WebUIPort)
		}
		go srv.Start()
		go s.Run()
		syncer.WaitForSignal(s)
		return
	}

	if err := app.Run(cfg, mode); err != nil {
		fmt.Fprintf(os.Stderr, "运行失败: %v\n", err)
		os.Exit(1)
	}
}

// getDataDir 获取数据存储目录
func getDataDir() string {
	// 优先使用环境变量（Docker 部署场景）
	if dir := os.Getenv("FWALIZER_DATA_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法获取用户目录: %v\n", err)
		os.Exit(1)
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "fwalizer")
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appdata, "fwalizer")
	default:
		return filepath.Join(home, ".config", "fwalizer")
	}
}
