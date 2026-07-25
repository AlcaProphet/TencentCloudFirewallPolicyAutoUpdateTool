package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alcaprophet/fwalizer/app"
	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/dns"
	"github.com/alcaprophet/fwalizer/provider"
	"github.com/alcaprophet/fwalizer/syncer"
	"github.com/alcaprophet/fwalizer/webui"
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

		// 初始化日志
		app.InitLogger(cfg.LogLevel)

		// 启动 WebUI 服务器
		srv := webui.NewServer(store, cfg.WebUIPort)

		// 如果有目标和规则，启动同步引擎并接通热重载
		if len(cfg.Targets) > 0 && len(cfg.DomainRules) > 0 {
			provider.SetCredentials(cfg.TCAccessID, cfg.TCAccessKey, cfg.AliAccessID, cfg.AliAccessKey)
			pool := provider.NewClientPool()
			var providers []provider.Provider
			for i, t := range cfg.Targets {
				p, err := provider.NewProvider(t, i, pool)
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

			// 接通热重载：WebUI 修改配置后重新加载并通知 Syncer
			srv.SetReloadFunc(func() {
				newCfg, err := store.LoadConfig()
				if err != nil {
					slog.Error("重载配置失败", "error", err)
					return
				}
				s.Reload(newCfg)
			})

			go srv.Start()
			go s.Run()
			syncer.WaitForSignal(s)
		} else {
			// 无配置时仅运行 WebUI，等待用户配置
			slog.Info("WebUI 已启动，请通过浏览器配置", "port", cfg.WebUIPort)
			if err := srv.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "WebUI 服务器失败: %v\n", err)
				os.Exit(1)
			}
		}
		return
	}

	if err := app.Run(cfg, mode); err != nil {
		fmt.Fprintf(os.Stderr, "运行失败: %v\n", err)
		os.Exit(1)
	}
}

// getDataDir 获取数据存储目录
func getDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法获取用户目录: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "fwalizer")
}
