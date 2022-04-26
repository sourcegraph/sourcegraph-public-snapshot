package log

import (
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/globallogger"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

type TraceContext = otfields.TraceContext

// Logger is an OpenTelemetry-compliant logger. All functions that log output should hold
// a reference to a Logger that gets passed in from callers, so as to maintain fields and
// context.
type Logger interface {
	// Scoped creates a new Logger with scope attached as part of its instrumentation
	// scope. For example, if the underlying logger is scoped 'foo', then
	// 'logger.Scoped("bar")' will create a logger with scope 'foo.bar'.
	//
	// Scopes should be static values, NOT dynamic values like identifiers or parameters.
	//
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-instrumentationscope
	Scoped(scope string, description string) Logger

	// With creates a new Logger with the given fields as attributes.
	//
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-attributes
	With(...Field) Logger
	// WithTrace creates a new Logger with the given trace context.
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
}

// Scoped returns the global logger and sets it up with the given scope and OpenTelemetry
// compliant implementation. Instead of using this everywhere a log is needed, callers
// should hold a reference to the Logger and pass it in to places that need to log.
//
// Scopes should be static values, NOT dynamic values like identifiers or parameters.
func Scoped(scope string, description string) Logger {
	safeGet := !development // do not panic in prod
	adapted := &zapAdapter{Logger: globallogger.Get(safeGet)}

	return adapted.Scoped(scope, description).With(otfields.AttributesNamespace)
}

type zapAdapter struct {
	*zap.Logger

	// scope is a read-only copy of this logger's full scope so that we can rebuild
	// loggers from root.
	scope string

	// attributes is a read-only copy of all attributes used in this logger, for the
	// purposes of being able to rebuild loggers from a root logger to bypass the
	// Attributes namespace.
	attributes []Field

	// options preserves options from initLogger, for a similar purpose to attributes
	// and scope.
	options []zap.Option
}

var _ Logger = &zapAdapter{}

// createdScopes tracks the scopes that have been created so far.
var createdScopes sync.Map

func (z *zapAdapter) Scoped(scope string, description string) Logger {
	var newScope string
	if z.scope == "" {
		newScope = scope
	} else {
		newScope = strings.Join([]string{z.scope, scope}, ".")
	}
	scopedLogger := &zapAdapter{
		Logger:     z.Logger.Named(scope), // name -> scope in OT
		scope:      newScope,
		attributes: z.attributes,
		options:    z.options,
	}
	if len(description) > 0 {
		if _, alreadyLogged := createdScopes.LoadOrStore(newScope, struct{}{}); !alreadyLogged {
			scopedLogger.Debug("logger.scoped",
				zap.String("scope", scope),
				zap.String("description", description))
		}
	}
	return scopedLogger
}

func (z *zapAdapter) With(fields ...Field) Logger {
	return &zapAdapter{
		Logger:     z.Logger.With(fields...),
		scope:      z.scope,
		attributes: append(z.attributes, fields...),
		options:    z.options,
	}
}

func (z *zapAdapter) WithTrace(trace TraceContext) Logger {
	newLogger := globallogger.Get(development).
		Named(z.scope).
		With(zap.Inline(&encoders.TraceContextEncoder{TraceContext: trace})).
		With(z.attributes...)
	if len(z.options) > 0 {
		newLogger = newLogger.WithOptions(z.options...)
	}
	return &zapAdapter{
		Logger:     newLogger,
		scope:      z.scope,
		attributes: z.attributes,
		options:    z.options,
	}
}

// WithOptions is an internal API used to allow packages like logtest to hook into the
// underlying zap logger.
func (z *zapAdapter) WithOptions(options ...zap.Option) Logger {
	return &zapAdapter{
		Logger:     z.Logger.WithOptions(options...),
		scope:      z.scope,
		attributes: z.attributes,
		options:    append(z.options, options...),
	}
}
