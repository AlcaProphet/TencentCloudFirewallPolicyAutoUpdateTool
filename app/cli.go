package app

import (
	"fmt"
	"os"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/version"
)

// RunCLI 处理子命令，返回 true 表示已处理（不需进入主流程）
func RunCLI(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch args[1] {
	case "version":
		fmt.Printf("fwalizer %s\n", version.Version)
		return true
	case "validate":
		path := ".env"
		if len(args) >= 3 {
			path = args[2]
		}
		cfg, err := config.LoadEnv(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "解析失败: %v\n", err)
			os.Exit(1)
		}
		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "校验失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("配置有效: %d 个目标, %d 条规则\n", len(cfg.Targets), len(cfg.DomainRules))
		return true
	}
	return false
}
