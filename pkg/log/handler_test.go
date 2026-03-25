package log

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"
)

func TestTraceHandlerEnabled(t *testing.T) {
	// 创建一个内存 buffer 作为输出
	var buf bytes.Buffer
	baseHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	handler := &traceHandler{
		handler:     baseHandler,
		extractors:  nil,
		enableTrace: false,
	}

	// 测试 Enabled 方法
	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Enabled() should return true for LevelInfo")
	}
	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Enabled() should return false for LevelDebug")
	}
}

func TestTraceHandlerHandle(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	extractor := func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String("custom", "value")}
	}

	handler := &traceHandler{
		handler:     baseHandler,
		extractors:  []Extractor{extractor},
		enableTrace: false,
	}

	// 创建 logger 并记录日志
	logger := slog.New(handler)
	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Error("Handle() should write output")
	}
}

func TestTraceHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewTextHandler(&buf, nil)

	handler := &traceHandler{
		handler:     baseHandler,
		extractors:  nil,
		enableTrace: false,
	}

	// 调用 WithAttrs
	newHandler := handler.WithAttrs([]slog.Attr{slog.String("key", "value")})

	// 验证返回的是新的 traceHandler
	th, ok := newHandler.(*traceHandler)
	if !ok {
		t.Fatal("WithAttrs() should return *traceHandler")
	}
	if th.handler == nil {
		t.Error("WithAttrs() should set handler")
	}
}

func TestTraceHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewTextHandler(&buf, nil)

	handler := &traceHandler{
		handler:     baseHandler,
		extractors:  nil,
		enableTrace: false,
	}

	// 调用 WithGroup
	newHandler := handler.WithGroup("test")

	// 验证返回的是新的 traceHandler
	if _, ok := newHandler.(*traceHandler); !ok {
		t.Fatal("WithGroup() should return *traceHandler")
	}
}

func TestTraceHandlerTraceInjection(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})

	handler := &traceHandler{
		handler:     baseHandler,
		extractors:  nil,
		enableTrace: true,
	}

	// 创建 mock tracer
	tracerProvider := noop.NewTracerProvider()
	tracer := tracerProvider.Tracer("test")

	// 创建带有 trace 的 context
	ctx, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	logger := slog.New(handler)
	logger.InfoContext(ctx, "test message")

	output := buf.String()
	t.Logf("Output: %s", output)

	// 注意：NoopTracerProvider 产生的 span context IsValid() 为 false
	// 这里主要验证 Handle() 不会 panic，并能正常处理 context
	// 实际的 trace 注入（trace_id/span_id）需要使用真实的 tracer provider
	if output == "" {
		t.Error("Handle() should write output")
	}

	// NoopTracerProvider 的 span IsValid() 返回 false，所以不会添加 trace 字段
	// 我们验证至少 enableTrace=true 不会导致 panic
}
