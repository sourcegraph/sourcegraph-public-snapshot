package log

import (
	"go.uber.org/zap"
)

type TraceContext struct {
	TraceID string
	SpanID  string
}

type Logger interface {
	// Named creates a new Logger with segment attached to the name.
	Named(segment string) Logger
	// With creates a new Logger with the given fields.
	With(...Field) Logger
	// WithTrace creates a new Logger with the given trace context.
	WithTrace(TraceContext) Logger

	// Debug logs a debug message, including any fields accumulated on the Logger.
	//
	// Debug logs are typically voluminous, and are usually disabled in production.
	Debug(string, ...Field)
	// Info logs an info message, including any fields accumulated on the Logger.
	//
	// Info is the default logging priority.
	Info(string, ...Field)
	// Warn logs a message at WarnLevel, including any fields accumulated on the Logger.
	//
	// Warning logs are more important than Info, but don't need individual human review.
	Warn(string, ...Field)
	// Error logs an error message, including any fields accumulated on the Logger.
	//
	// Error logs are high-priority. If an application is running smoothly, it shouldn't
	// generate any error-level logs.
	Error(string, ...Field)
	// Critical logs a critical message, including any fields accumulated on the
	// Logger. In development mode, this also causes a panic.
	//
	// Critical logs are particularly important logs.
	Critical(string, ...Field)

	// Sync flushes any buffered log entries. Applications should take care to call Sync
	// before exiting.
	Sync() error
}

type zapAdapter struct {
	*zap.Logger

	// allFields is a read-only copy of all allFields used in this logger, for the
	// purposes of being able to rebuild loggers from a root logger to bypass the
	// Attributes namespace.
	allFields []Field

	// options preserves options from initLogger, for a similar purpose to allFields
	options []zap.Option
}

var _ Logger = &zapAdapter{}

func (z *zapAdapter) Named(segment string) Logger {
	return &zapAdapter{
		Logger:    z.Logger.Named(segment),
		allFields: z.allFields,
		options:   z.options,
	}
}

func (z *zapAdapter) With(fields ...Field) Logger {
	return &zapAdapter{
		Logger:    z.Logger.With(fields...),
		allFields: append(z.allFields, fields...),
		options:   z.options,
	}
}

func (z *zapAdapter) WithTrace(trace TraceContext) Logger {
	allFields := append([]Field{zap.Inline(&traceContext{trace})}, z.allFields...)
	logger := getGlobal().Logger.With(allFields...)
	if len(z.options) > 0 {
		logger = logger.WithOptions(z.options...)
	}
	return &zapAdapter{
		Logger:    logger,
		allFields: allFields,
		options:   z.options,
	}
}

func (z *zapAdapter) Critical(msg string, fields ...Field) {
	z.Logger.DPanic(msg, fields...)
}

func (z *zapAdapter) withOptions(options ...zap.Option) Logger {
	return &zapAdapter{
		Logger:    z.Logger.WithOptions(options...),
		allFields: z.allFields,
		options:   append(z.options, options...),
	}
}
