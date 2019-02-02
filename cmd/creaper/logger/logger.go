package logger

import (
	"context"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

type key int

const (
	loggerKey key = iota
)

func WithLogger(ctx context.Context, logCtx ...interface{}) context.Context {
	return context.WithValue(
		ctx,
		loggerKey,
		Logger(ctx).New(logCtx...),
	)
}

func Logger(ctx context.Context) log15.Logger {
	if existing, found := ctx.Value(loggerKey).(log15.Logger); found {
		return existing
	}

	return log15.Root()
}

func Debug(ctx context.Context, msg string, logCtx ...interface{}) {
	Logger(ctx).Debug(msg, logCtx...)
}

func Info(ctx context.Context, msg string, logCtx ...interface{}) {
	Logger(ctx).Info(msg, logCtx...)
}

func Warn(ctx context.Context, msg string, logCtx ...interface{}) {
	Logger(ctx).Warn(msg, logCtx...)
}

func Error(ctx context.Context, msg string, logCtx ...interface{}) {
	Logger(ctx).Error(msg, logCtx...)
}

func Crit(ctx context.Context, msg string, logCtx ...interface{}) {
	Logger(ctx).Crit(msg, logCtx...)
}
