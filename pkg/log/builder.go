package log

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Builder 日志构建器
type Builder struct {
	config     Config
	extractors []Extractor
}

// NewBuilder 创建新的 Builder
func NewBuilder() *Builder {
	return &Builder{
		config: Config{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		extractors: make([]Extractor, 0),
	}
}

// FromConfig 从配置加载
func (b *Builder) FromConfig(cfg *Config) *Builder {
	b.config = *cfg
	return b
}

// WithLevel 设置日志级别
func (b *Builder) WithLevel(level string) *Builder {
	b.config.Level = level
	return b
}

// WithFormat 设置输出格式
func (b *Builder) WithFormat(format string) *Builder {
	b.config.Format = format
	return b
}

// WithOutput 设置输出位置
func (b *Builder) WithOutput(output string) *Builder {
	b.config.Output = output
	return b
}

// WithTraceEnabled 设置是否启用 trace
func (b *Builder) WithTraceEnabled(enabled bool) *Builder {
	b.config.EnableTrace = enabled
	return b
}

// WithAddSource 设置是否添加源码位置
func (b *Builder) WithAddSource(add bool) *Builder {
	b.config.AddSource = add
	return b
}

// WithTimeFormat 设置时间格式
func (b *Builder) WithTimeFormat(format string) *Builder {
	b.config.TimeFormat = format
	return b
}

// WithExtractor 添加字段提取器
func (b *Builder) WithExtractor(fn Extractor) *Builder {
	b.extractors = append(b.extractors, fn)
	return b
}

// parseLevel 解析日志级别字符串（不区分大小写）
func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown level: %q (valid: debug, info, warn, error)", level)
	}
}

// createWriter 根据输出配置创建 Writer
func createWriter(output string) (io.Writer, error) {
	switch output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("output must be 'stdout' or 'stderr', got %q", output)
	}
}

// createAttrReplacer 创建属性替换函数（处理时间格式）
func createAttrReplacer(timeFormat string) func([]string, slog.Attr) slog.Attr {
	if timeFormat == "" {
		timeFormat = "2006-01-02 15:04:05"
	}
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "time" {
			if t, ok := a.Value.Any().(time.Time); ok {
				return slog.String("time", t.Format(timeFormat))
			}
		}
		return a
	}
}

// Build 构建 logger
func (b *Builder) Build() (*slog.Logger, error) {
	// 1. 验证配置
	if err := b.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 2. 解析日志级别
	level, err := parseLevel(b.config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// 3. 创建 Writer
	writer, err := createWriter(b.config.Output)
	if err != nil {
		return nil, fmt.Errorf("create output writer: %w", err)
	}

	// 4. 创建 HandlerOptions
	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   b.config.AddSource,
		ReplaceAttr: createAttrReplacer(b.config.TimeFormat),
	}

	// 5. 创建基础 Handler
	var baseHandler slog.Handler
	if b.config.Format == "json" {
		baseHandler = slog.NewJSONHandler(writer, opts)
	} else {
		baseHandler = slog.NewTextHandler(writer, opts)
	}

	// 6. 包装 traceHandler
	handler := &traceHandler{
		handler:     baseHandler,
		extractors:  b.extractors,
		enableTrace: b.config.EnableTrace,
	}

	// 7. 创建并返回 Logger
	return slog.New(handler), nil
}

// BuildAsDefault 构建并设置为 slog 默认 logger
// 返回 logger 和恢复函数
func (b *Builder) BuildAsDefault() (*slog.Logger, func(), error) {
	logger, err := b.Build()
	if err != nil {
		return nil, nil, err
	}

	// 保存当前默认 logger
	oldDefault := slog.Default()

	// 设置新的默认 logger
	slog.SetDefault(logger)

	// 返回恢复函数
	resetFunc := func() {
		slog.SetDefault(oldDefault)
	}

	return logger, resetFunc, nil
}
