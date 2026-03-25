package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yanking/price-watch/cmd/watch/initial"
	"github.com/yanking/price-watch/internal/watch/config"
	"github.com/yanking/price-watch/internal/watch/svc"
	"github.com/yanking/price-watch/pkg/app"
	"github.com/yanking/price-watch/pkg/conf"
)

// 构建时通过 -ldflags 注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func main() {
	// 检查 --version 参数（flag 包不支持长选项，需手动处理）
	for _, arg := range os.Args[1:] {
		if arg == "--version" {
			fmt.Printf("Version:    %s\n", Version)
			fmt.Printf("Build Time: %s\n", BuildTime)
			fmt.Printf("Go Version: %s\n", GoVersion)
			os.Exit(0)
		}
	}

	configFile := flag.String("c", "./configs/watch.yaml", "path to config file.")
	showVersion := flag.Bool("v", false, "show version information.")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Go Version: %s\n", GoVersion)
		os.Exit(0)
	}

	// 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create service context: %v\n", err)
		os.Exit(1)
	}
	// 添加 defer 关闭
	defer func() {
		if cErr := ctx.Close(); cErr != nil {
			ctx.Logger.Error("close service context", "error", cErr)
		}
	}()

	// 初始化应用
	initial.App(ctx)

	// 创建服务 & 启动服务
	servers := initial.CreateServices(ctx)
	closes := initial.Close(servers)
	a := app.New(servers, closes)
	if aErr := a.Run(); aErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "application error: %v\n", aErr)
		os.Exit(1)
	}
}
