package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// Server 表示一个可以被应用管理的服务器。
type Server interface {
	Start() error
	Stop() error
	String() string
}

// CleanupFunc 是一个清理函数，用于释放应用所占用的资源。
type CleanupFunc func() error

// App 表示一个应用，管理一组服务器和清理函数。
type App struct {
	servers  []Server
	cleanups []CleanupFunc
}

// New 创建一个新的应用实例。
func New(servers []Server, cleanups []CleanupFunc) *App {
	return &App{
		servers:  servers,
		cleanups: cleanups,
	}
}

// Run 启动应用并阻塞直到应用停止。
func (a *App) Run() error {
	eg, ctx := errgroup.WithContext(context.Background())

	for _, server := range a.servers {
		_s := server
		eg.Go(func() error {
			fmt.Printf("starting %s...\n", _s.String())
			return _s.Start()
		})
	}

	eg.Go(func() error {
		return a.watch(ctx)
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("app run error: %w", err)
	}

	return nil
}

// watch 监听上下文取消和系统信号，以优雅地停止应用。
func (a *App) watch(ctx context.Context) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		select {
		case <-ctx.Done():
			if err := a.stop(); err != nil {
				return fmt.Errorf("app stopped due to context cancellation: %w", err)
			}
			return fmt.Errorf("app stopped due to context cancellation: %w", ctx.Err())
		case sigType := <-sig:
			fmt.Printf("received system notification signal: %s, stopping...\n", sigType.String())
			switch sigType {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP:
				if err := a.stop(); err != nil {
					return fmt.Errorf("app stopped due to signal %s: %w", sigType.String(), err)
				}
				fmt.Println("stop app successfully")
				return nil
			}
		}
	}
}

// stop 执行所有清理函数。
func (a *App) stop() error {
	for _, cleanup := range a.cleanups {
		if err := cleanup(); err != nil {
			return fmt.Errorf("close app error: %w", err)
		}
	}

	return nil
}
