package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// traceHandler 包装标准 Handler，集成字段提取和 trace 注入
type traceHandler struct {
	handler     slog.Handler
	extractors  []Extractor
	enableTrace bool
}

// Enabled 实现 slog.Handler 接口
func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. 提取 trace 信息（如果启用）
	if h.enableTrace {
		span := trace.SpanFromContext(ctx)
		if span.SpanContext().IsValid() {
			r.AddAttrs(
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			)
		}
	}

	// 2. 执行自定义提取器（带 panic 保护）
	for _, extractor := range h.extractors {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// 记录到 stderr（避免递归调用 logger）
					fmt.Fprintf(os.Stderr, "log extractor panic: %v\n", r)
				}
			}()
			if attrs := extractor(ctx); len(attrs) > 0 {
				r.AddAttrs(attrs...)
			}
		}()
	}

	// 3. 委托给底层 Handler
	return h.handler.Handle(ctx, r)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{
		handler:     h.handler.WithAttrs(attrs),
		extractors:  h.extractors,
		enableTrace: h.enableTrace,
	}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{
		handler:     h.handler.WithGroup(name),
		extractors:  h.extractors,
		enableTrace: h.enableTrace,
	}
}
