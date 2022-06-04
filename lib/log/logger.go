package log

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

	// AddCallerSkip increases the number of callers skipped by caller annotation. When
	// building wrappers around the Logger, supplying this Option prevents the Logger from
	// always reporting the wrapper code as the caller.
	AddCallerSkip(int) Logger
	// IncreaseLevel creates a logger that only logs at or above the given level for the given
	// scope. To disable all output, you can use LogLevelNone.
	//
	// IncreaseLevel is only allowed to increase the level the Logger was initialized at -
	// it has no affect if the preset level is higher than the inidcated level.
	IncreaseLevel(scope string, description string, level Level) Logger
}

// Scoped returns the global logger and sets it up with the given scope and OpenTelemetry
// compliant implementation. Instead of using this everywhere a log is needed, callers
// should hold a reference to the Logger and pass it in to places that need to log.
//
// Scopes should be static values, NOT dynamic values like identifiers or parameters.
func Scoped(scope string, description string) Logger {
	devMode := globallogger.DevMode()
	safeGet := !devMode // do not panic in prod
	root := globallogger.Get(safeGet)
	adapted := &zapAdapter{
		Logger:            root,
		rootLogger:        root,
		fromPackageScoped: true,
	}

	if devMode {
		// In development, don't add the OpenTelemetry "Attributes" namespace which gets
		// rather difficult to read.
		return adapted.Scoped(scope, description)
	}
	return adapted.Scoped(scope, description).With(otfields.AttributesNamespace)
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

// createdScopes tracks the scopes that have been created so far.
var createdScopes sync.Map

func (z *zapAdapter) Scoped(scope string, description string) Logger {
	var newFullScope string
	if z.fullScope == "" {
		newFullScope = scope
	} else {
		newFullScope = createScope(z.fullScope, scope)
	}
	scopedLogger := &zapAdapter{
		// name -> scope in OT
		Logger:     z.Logger.Named(scope),
		rootLogger: z.rootLogger.Named(scope),

		fullScope:  newFullScope,
		attributes: z.attributes,
	}
	if len(description) > 0 {
		if _, alreadyLogged := createdScopes.LoadOrStore(newFullScope, struct{}{}); !alreadyLogged {
			callerSkip := 1 // Logger.Scoped() -> Logger.Debug()
			if z.fromPackageScoped {
				callerSkip += 1 // log.Scoped() -> Logger.Scoped() -> Logger.Debug()
			}
			scopedLogger.
				AddCallerSkip(callerSkip).
				Debug("logger.scoped",
					zap.String("scope", scope),
					zap.String("description", description))
		}
	}
	return scopedLogger
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
// It must implement logtest.configurableAdapter
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
