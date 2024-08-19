package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/log/internal/encoders"
	"github.com/sourcegraph/log/internal/globallogger"
	"github.com/sourcegraph/log/internal/otelfields"
)

// TraceContext represents a trace to associate with log entries.
//
// https://opentelemetry.io/docs/reference/specification/logs/data-model/#trace-context-fields
type TraceContext = otelfields.TraceContext

// Logger is an OpenTelemetry-compliant logger. All functions that log output should hold
// a reference to a Logger that gets passed in from callers, so as to maintain fields and
// context.
type Logger interface {
	// Scoped creates a new Logger with scope attached as part of its instrumentation
	// scope. For example, if the underlying logger is scoped 'foo', then
	// 'logger.Scoped("bar")' will create a logger with scope 'foo.bar'.
	//
	// Scopes should be static values, NOT dynamic values like identifiers or parameters,
	// and should generally be CamelCased with descriptions that follow our logging
	// conventions - learn more: https://docs.sourcegraph.com/dev/how-to/add_logging#scoped-loggers
	//
	// Scopes map to OpenTelemetry InstrumentationScopes:
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-instrumentationscope
	Scoped(scope string) Logger

	// With creates a new Logger with the given fields as attributes.
	//
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-attributes
	With(...Field) Logger
	// WithTrace creates a new Logger with the given trace context. If TraceContext has no
	// fields set, this function is a no-op. If WithTrace has already been called on this
	// logger with a valid TraceContext, the existing TraceContext will be overwritten
	// with the new TraceContext.
	//
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#trace-context-fields
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
	// Fatal logs a fatal error message, including any fields accumulated on the Logger.
	// The logger then calls os.Exit(1), flushing the logger before doing so. Use sparingly.
	Fatal(string, ...Field)

	// AddCallerSkip increases the number of callers skipped by caller annotation. When
	// building wrappers around the Logger, using AddCallerSkip prevents the Logger from
	// always reporting the wrapper code as the caller.
	AddCallerSkip(int) Logger
	// IncreaseLevel creates a logger that only logs at or above the given level for the given
	// scope. To disable all output, you can use LogLevelNone.
	//
	// IncreaseLevel is only allowed to increase the level the Logger was initialized at -
	// it has no affect if the preset level is higher than the indicated level.
	IncreaseLevel(scope string, description string, level Level) Logger
}

// Scoped returns the global logger and sets it up with the given scope and OpenTelemetry
// compliant implementation. Instead of using this everywhere a log is needed, callers
// should hold a reference to the Logger and pass it in to places that need to log.
//
// Scopes should be static values, NOT dynamic values like identifiers or parameters,
// and should generally be CamelCased with descriptions that follow our logging
// conventions - learn more: https://docs.sourcegraph.com/dev/how-to/add_logging#scoped-loggers
//
// When testing, you should use 'logtest.Scoped(*testing.T)' instead - learn more:
// https://docs.sourcegraph.com/dev/how-to/add_logging#testing-usage
func Scoped(scope string) Logger {
	safeGet := !globallogger.DevMode() // do not panic in prod
	root := globallogger.Get(safeGet)
	adapted := &zapAdapter{
		Logger:            root,
		rootLogger:        root,
		fromPackageScoped: true,
	}

	if globallogger.DevMode() {
		// In development, don't add the OpenTelemetry "Attributes" namespace which gets
		// rather difficult to read.
		return adapted.Scoped(scope)
	}
	return adapted.Scoped(scope).With(otelfields.AttributesNamespace)
}

// NoOp returns a no-op Logger that can never produce any output. It can be safely created
// before initialization. Use sparingly, and do not use with the intent of replacing it
// post-initialization.
func NoOp() Logger {
	root := zap.NewNop()
	return &zapAdapter{
		Logger:     root,
		rootLogger: root,
	}
}

type zapAdapter struct {
	*zap.Logger

	// rootLogger is used to rebuild loggers with fields that bypass the Attributes
	// namespace. Keep this in sync with all modifications made to Logger.
	rootLogger *zap.Logger

	// fullScope tracks the full name of the logger's scope, for logging scope descriptions.
	fullScope string

	// attributes is a read-only copy of all attributes used in this logger, for the
	// purposes of being able to rebuild loggers from a root logger to bypass the
	// Attributes namespace.
	attributes []Field

	// fromPackageScoped indicates this logger is from log.Scoped. Do not copy this to
	// child loggers, and do not set this anywhere except log.Scoped.
	fromPackageScoped bool
}

var _ Logger = &zapAdapter{}

func (z *zapAdapter) Scoped(scope string) Logger {
	var newFullScope string
	if z.fullScope == "" {
		newFullScope = scope
	} else {
		newFullScope = createScope(z.fullScope, scope)
	}
	return &zapAdapter{
		// name -> scope in OT
		Logger:     z.Logger.Named(scope),
		rootLogger: z.rootLogger.Named(scope),

		fullScope:  newFullScope,
		attributes: z.attributes,
	}
}

func (z *zapAdapter) With(fields ...Field) Logger {
	return &zapAdapter{
		Logger:     z.Logger.With(fields...),
		rootLogger: z.rootLogger,
		fullScope:  z.fullScope,
		attributes: append(z.attributes, fields...),
	}
}

func (z *zapAdapter) WithTrace(trace TraceContext) Logger {
	if trace.TraceID == "" && trace.SpanID == "" {
		return z // no-op
	}

	// Reconstruct the logger - the TraceContext is not added to z.attributes, so this
	// effectively overwrites any existing TraceContext set with the new one. Note that
	// we never get to this point with a zero-value TraceContext, which no-ops earlier.
	newLogger := z.rootLogger.
		// insert trace before attributes
		With(zap.Inline(&encoders.TraceContextEncoder{TraceContext: trace})).
		// add attributes back
		With(z.attributes...)

	return &zapAdapter{
		Logger:     newLogger,
		rootLogger: z.rootLogger,
		fullScope:  z.fullScope,
		attributes: z.attributes,
	}
}

func (z *zapAdapter) AddCallerSkip(skip int) Logger {
	return &zapAdapter{
		Logger:     z.Logger.WithOptions(zap.AddCallerSkip(skip)),
		rootLogger: z.rootLogger.WithOptions(zap.AddCallerSkip(skip)),
		fullScope:  z.fullScope,
		attributes: z.attributes,
	}
}

func (z *zapAdapter) IncreaseLevel(scope string, description string, level Level) Logger {
	z.AddCallerSkip(1).Debug("logger.IncreaseLevel",
		Object("scope",
			String("scope", createScope(z.fullScope, scope)),
			String("description", description)),
		String("level", string(level)))

	opt := zap.IncreaseLevel(level.Parse())
	return &zapAdapter{
		Logger:     z.Logger.WithOptions(opt),
		rootLogger: z.rootLogger.WithOptions(opt),
		fullScope:  z.fullScope,
		attributes: z.attributes,
	}
}

// WithCore is an internal API used to allow packages like logtest to hook into
// underlying zap logger's core.
//
// It must implement internal/configurable.Logger. We do not assert it, however, because
// that would cause an import cycle - instead, there is a test in package configurable.
func (z *zapAdapter) WithCore(f func(c zapcore.Core) zapcore.Core) Logger {
	newRootLogger := z.rootLogger.
		WithOptions(zap.WrapCore(f))

	newLogger := newRootLogger.
		// add fields back
		With(z.attributes...)

	return &zapAdapter{
		Logger:     newLogger,
		rootLogger: newRootLogger,
		fullScope:  z.fullScope,
		attributes: z.attributes,
	}
}

func createScope(parent, child string) string {
	return fmt.Sprintf("%s.%s", parent, child)
}
