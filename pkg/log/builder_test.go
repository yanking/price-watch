package log

import (
	"context"
	"log/slog"
	"testing"
)

// TestNewBuilder 测试 NewBuilder 创建默认配置
func TestNewBuilder(t *testing.T) {
	b := NewBuilder()

	if b.config.Level != "info" {
		t.Errorf("expected default level 'info', got %q", b.config.Level)
	}
	if b.config.Format != "text" {
		t.Errorf("expected default format 'text', got %q", b.config.Format)
	}
	if b.config.Output != "stdout" {
		t.Errorf("expected default output 'stdout', got %q", b.config.Output)
	}
	if len(b.extractors) != 0 {
		t.Errorf("expected empty extractors, got %d", len(b.extractors))
	}
}

// TestBuilderChaining 测试链式调用
func TestBuilderChaining(t *testing.T) {
	extractor := func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String("test", "value")}
	}

	b := NewBuilder().
		WithLevel("debug").
		WithFormat("json").
		WithOutput("stderr").
		WithTraceEnabled(true).
		WithAddSource(true).
		WithTimeFormat("2006-01-02").
		WithExtractor(extractor)

	if b.config.Level != "debug" {
		t.Errorf("expected level 'debug', got %q", b.config.Level)
	}
	if b.config.Format != "json" {
		t.Errorf("expected format 'json', got %q", b.config.Format)
	}
	if b.config.Output != "stderr" {
		t.Errorf("expected output 'stderr', got %q", b.config.Output)
	}
	if !b.config.EnableTrace {
		t.Error("expected EnableTrace true")
	}
	if !b.config.AddSource {
		t.Error("expected AddSource true")
	}
	if b.config.TimeFormat != "2006-01-02" {
		t.Errorf("expected TimeFormat '2006-01-02', got %q", b.config.TimeFormat)
	}
	if len(b.extractors) != 1 {
		t.Errorf("expected 1 extractor, got %d", len(b.extractors))
	}
}

// TestBuilderWithExtractor 测试添加提取器
func TestBuilderWithExtractor(t *testing.T) {
	b := NewBuilder()

	extractor1 := func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String("key1", "value1")}
	}
	extractor2 := func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String("key2", "value2")}
	}

	b.WithExtractor(extractor1).WithExtractor(extractor2)

	if len(b.extractors) != 2 {
		t.Fatalf("expected 2 extractors, got %d", len(b.extractors))
	}
}

// TestParseLevel 测试日志级别解析
func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    slog.Level
		wantErr bool
	}{
		{"debug lower", "debug", slog.LevelDebug, false},
		{"DEBUG upper", "DEBUG", slog.LevelDebug, false},
		{"info lower", "info", slog.LevelInfo, false},
		{"INFO upper", "INFO", slog.LevelInfo, false},
		{"warn", "warn", slog.LevelWarn, false},
		{"warning", "warning", slog.LevelWarn, false},
		{"WARN upper", "WARN", slog.LevelWarn, false},
		{"error", "error", slog.LevelError, false},
		{"ERROR upper", "ERROR", slog.LevelError, false},
		{"invalid", "invalid", slog.LevelInfo, true},
		{"empty", "", slog.LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCreateWriter 测试 Writer 创建
func TestCreateWriter(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{"stdout", "stdout", false},
		{"stderr", "stderr", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := createWriter(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("createWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if writer != nil {
					t.Error("createWriter() should return nil on error")
				}
				return
			}
			if writer == nil {
				t.Error("createWriter() writer should not be nil")
			}
		})
	}
}

// TestBuilderBuild 测试 Build 方法
func TestBuilderBuild(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Builder)
		wantErr bool
	}{
		{
			name:    "default config",
			setup:   func(b *Builder) {},
			wantErr: false,
		},
		{
			name: "with all options",
			setup: func(b *Builder) {
				extractor := func(ctx context.Context) []slog.Attr {
					return []slog.Attr{slog.String("custom", "value")}
				}
				b.WithLevel("debug").
					WithFormat("json").
					WithOutput("stderr").
					WithTraceEnabled(true).
					WithAddSource(true).
					WithTimeFormat("2006-01-02").
					WithExtractor(extractor)
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			setup: func(b *Builder) {
				b.WithFormat("invalid")
			},
			wantErr: true,
		},
		{
			name: "invalid level",
			setup: func(b *Builder) {
				b.WithLevel("invalid")
			},
			wantErr: true,
		},
		{
			name: "empty output",
			setup: func(b *Builder) {
				b.WithOutput("")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			tt.setup(b)

			logger, err := b.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if logger != nil {
					t.Error("Build() should return nil logger on error")
				}
				return
			}
			if logger == nil {
				t.Error("Build() should return non-nil logger")
			}

			// 验证 logger 可以正常工作
			// 实际日志输出测试在集成测试中更合适
		})
	}
}

// TestBuilderBuildAsDefault 测试 BuildAsDefault 方法
func TestBuilderBuildAsDefault(t *testing.T) {
	// 保存原始的默认 logger
	oldDefault := slog.Default()
	defer slog.SetDefault(oldDefault)

	b := NewBuilder().
		WithOutput("stdout").
		WithLevel("info").
		WithFormat("text")

	logger, resetFunc, err := b.BuildAsDefault()
	if err != nil {
		t.Fatalf("BuildAsDefault() error = %v", err)
	}
	if logger == nil {
		t.Fatal("BuildAsDefault() should return non-nil logger")
	}
	if resetFunc == nil {
		t.Fatal("BuildAsDefault() should return non-nil reset function")
	}

	// 验证新的默认 logger
	newDefault := slog.Default()
	if newDefault == nil {
		t.Error("slog.Default() should return non-nil logger")
	}
	// 注意：由于我们使用的是 stdout，实际 logger 可能不同
	// 这里只验证 Default() 能正常工作

	// 测试恢复函数
	resetFunc()

	restoredDefault := slog.Default()
	// 恢复后应该回到原始 logger
	if restoredDefault != oldDefault {
		// slog 的 Default() 返回的是指针，所以可以直接比较
		t.Error("logger should be restored after calling resetFunc()")
	}
}

// TestFromConfig 测试 FromConfig 方法
func TestFromConfig(t *testing.T) {
	cfg := &Config{
		Level:       "debug",
		Format:      "json",
		Output:      "stderr",
		EnableTrace: true,
		AddSource:   true,
		TimeFormat:  "2006-01-02",
	}

	b := NewBuilder().FromConfig(cfg)

	if b.config.Level != "debug" {
		t.Errorf("expected level 'debug', got %q", b.config.Level)
	}
	if b.config.Format != "json" {
		t.Errorf("expected format 'json', got %q", b.config.Format)
	}
	if b.config.Output != "stderr" {
		t.Errorf("expected output 'stderr', got %q", b.config.Output)
	}
	if !b.config.EnableTrace {
		t.Error("expected EnableTrace true")
	}
	if !b.config.AddSource {
		t.Error("expected AddSource true")
	}
	if b.config.TimeFormat != "2006-01-02" {
		t.Errorf("expected TimeFormat '2006-01-02', got %q", b.config.TimeFormat)
	}
}
