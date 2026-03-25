package log_test

import (
	"context"
	"log/slog"

	"github.com/yanking/price-watch/pkg/log"
)

// contextKey 定义自定义 context key 类型，避免碰撞
type contextKey string

const userIDKey contextKey = "userID"

func ExampleBuilder_basic() {
	// 从配置文件构建
	cfg := log.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	builder := log.NewBuilder().FromConfig(&cfg)
	logger, err := builder.Build()
	if err != nil {
		panic(err)
	}

	logger.Info("message", slog.String("key", "value"))
}

func ExampleBuilder_withExtractor() {
	// 添加自定义提取器
	builder := log.NewBuilder().
		WithLevel("info").
		WithOutput("stdout").
		WithExtractor(func(ctx context.Context) []slog.Attr {
			if id, ok := ctx.Value(userIDKey).(string); ok {
				return []slog.Attr{slog.String("userID", id)}
			}
			return nil
		})

	logger, err := builder.Build()
	if err != nil {
		panic(err)
	}

	ctx := context.WithValue(context.Background(), userIDKey, "user123")
	logger.InfoContext(ctx, "user action")
}

func ExampleBuilder_asDefault() {
	// 设置为全局默认 logger
	builder := log.NewBuilder().
		WithLevel("info").
		WithOutput("stdout")

	logger, reset, err := builder.BuildAsDefault()
	if err != nil {
		panic(err)
	}
	defer reset()

	_ = logger

	// 使用 slog 顶级函数
	slog.Info("global message", "key", "value")

	// 使用 *Ctx 方法启用 trace
	ctx := context.Background()
	slog.InfoContext(ctx, "message with context")
}

func ExampleBuilder_withTrace() {
	// 启用 OpenTelemetry trace
	builder := log.NewBuilder().
		WithLevel("info").
		WithFormat("json").
		WithOutput("stdout").
		WithTraceEnabled(true).
		WithAddSource(true)

	logger, err := builder.Build()
	if err != nil {
		panic(err)
	}

	// 在带有 trace span 的 context 中使用
	ctx := context.Background()
	logger.InfoContext(ctx, "operation completed", slog.String("operation", "process"))
}

func ExampleBuilder_customTimeFormat() {
	// 自定义时间格式
	builder := log.NewBuilder().
		WithLevel("info").
		WithOutput("stdout").
		WithTimeFormat("2006-01-02 15:04:05")

	logger, err := builder.Build()
	if err != nil {
		panic(err)
	}

	logger.Info("message with custom time format")
}

func ExampleBuilder_stderr() {
	// 输出到 stderr
	builder := log.NewBuilder().
		WithLevel("debug").
		WithFormat("text").
		WithOutput("stderr")

	logger, err := builder.Build()
	if err != nil {
		panic(err)
	}

	logger.Debug("debugging info", slog.Int("count", 42))
}
