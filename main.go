package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/firewall"
)

func main() {
	// 初始化日志：JSON 格式输出到 stdout，供 docker logs 查看
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	slog.Info("FWAlizer 启动", "version", "0.1.0", "instance", cfg.InstanceID, "region", cfg.Region)

	// 2. 初始化同步器
	syncer, err := firewall.NewSyncer(cfg)
	if err != nil {
		slog.Error("初始化同步器失败", "error", err)
		os.Exit(1)
	}

	// 3. 监听退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// 4. 启动同步循环
	go syncer.Run()

	// 5. 等待退出信号
	sig := <-sigCh
	slog.Info("收到退出信号，正在优雅关闭...", "signal", sig.String())
	syncer.Shutdown()
	slog.Info("FWAlizer 已退出")
}
