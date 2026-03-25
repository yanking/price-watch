package log

import (
	"context"
	"log/slog"
)

// Extractor 字段提取器函数类型
// 从 context 中提取自定义日志字段
//
// 使用注意事项：
// - 提取器应该快速执行，避免耗时操作
// - 如果提取器可能 panic，traceHandler 会自动恢复并记录到 stderr
// - 返回 nil 或空切片表示没有提取到字段
type Extractor func(ctx context.Context) []slog.Attr
