package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alcaprophet/fwalizer/app"
	"github.com/alcaprophet/fwalizer/config"
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
		// WebUI 模式：SQLite + HTTP Server
		dataDir := getDataDir()
		os.MkdirAll(dataDir, 0755)
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

		// 启动 WebUI 服务器
		srv := webui.NewServer(store, cfg.WebUIPort)
		go func() {
			if err := srv.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "WebUI 服务器失败: %v\n", err)
				os.Exit(1)
			}
		}()

		// 如果有目标和规则，启动同步引擎
		if len(cfg.Targets) > 0 && len(cfg.DomainRules) > 0 {
			if err := app.Run(cfg, mode); err != nil {
				fmt.Fprintf(os.Stderr, "运行失败: %v\n", err)
				os.Exit(1)
			}
		} else {
			// 无配置时仅运行 WebUI，等待用户配置
			fmt.Println("WebUI 已启动，请通过浏览器配置")
			select {} // 阻塞等待
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
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fwalizer")
}
