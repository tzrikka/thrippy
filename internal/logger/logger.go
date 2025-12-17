package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

var ctxLoggerKey = ctxKey{}

func InContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if ctxLogger, ok := ctx.Value(ctxLoggerKey).(*slog.Logger); ok {
		l = ctxLogger
	}
	return l
}
