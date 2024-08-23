package log

import (
	"os"

	"github.com/sourcegraph/log/internal/globallogger"
	"github.com/sourcegraph/log/internal/otelfields"
)

var (
	// EnvDevelopment is key of the environment variable that is used to set whether
	// to use development logger configuration on Init.
	EnvDevelopment = globallogger.EnvDevelopment
	// EnvLogFormat is key of the environment variable that is used to set the log format
	// on Init.
	//
	// The value should be one of 'json', 'json_gcp' or 'condensed', defaulting to 'json'.
	EnvLogFormat = "SRC_LOG_FORMAT"
	// EnvLogLevel is key of the environment variable that can be used to set the log
	// level on Init.
	//
	// The value is one of 'debug', 'info', 'warn', 'error', or 'none', defaulting to
	// 'warn'.
	EnvLogLevel = "SRC_LOG_LEVEL"
	// EnvLogScopeLevel is key of the environment variable that can be used to
	// override the log level for specific scopes and its children.
	//
	// It has the format "SCOPE_0=LEVEL_0,SCOPE_1=LEVEL_1,...".
	//
	// Notes:
	//
	//  - these levels do not respect the root level (SRC_LOG_LEVEL), so this
	//    allows operators to turn up the verbosity of specific logs.
	//  - this only affects the outputcore (ie will not effect sentrycore).
	//  - Scope matches the full scope name. IE the below example has the scope
	//    "foo.bar" not "bar".
	//
	//    log.Scoped("foo", "").Scoped("bar", "")
	EnvLogScopeLevel = "SRC_LOG_SCOPE_LEVEL"
	// EnvLogSamplingInitial is key of the environment variable that can be used to set
	// the number of entries with identical messages to always output per second.
	//
	// Defaults to 100 - set explicitly to 0 or -1 to disable.
	EnvLogSamplingInitial = "SRC_LOG_SAMPLING_INITIAL"
	// EnvLogSamplingThereafter is key of the environment variable that can be used to set
	// the number of entries with identical messages to discard before emitting another
	// one per second, after EnvLogSamplingInitial.
	//
	// Defaults to 100 - set explicitly to 0 or -1 to disable.
	EnvLogSamplingThereafter = "SRC_LOG_SAMPLING_THEREAFTER"
)

type Resource = otelfields.Resource

// PostInitCallbacks is a set of callbacks returned by Init that enables finalization and
// updating of any configured sinks.
type PostInitCallbacks struct {
	// Sync must be called before application exit, such as via defer.
	//
	// Note: The error from sync is suppressed since this is usually called as a
	// defer in func main. In that case there isn't a reasonable way to handle the
	// error. As such this function signature doesn't return an error.
	Sync func()

	// Update should be called to change sink configuration, e.g. via
	// conf.Watch. Note that sinks not created upon initialization will
	// not be created post-initialization. Is a no-op if no sinks are enabled.
	Update func(SinksConfigGetter) func()
}

// Init initializes the log package's global logger as a logger of the given resource.
// It must be called on service startup, i.e. 'main()', NOT on an 'init()' function.
// Subsequent calls will panic, so do not call this within a non-service context.
//
// Init returns a set of callbacks - see PostInitCallbacks for more details. The Sync
// callback in particular must be called before application exit.
//
// For testing, you can use 'logtest.Init' to initialize the logging library.
//
// If Init is not called, trying to create a logger with Scoped will panic.
func Init(r Resource, s ...Sink) *PostInitCallbacks {
	if globallogger.IsInitialized() {
		panic("log.Init initialized multiple times")
	}

	// On initialization we get dev mode from env directly instead of from globallogger's
	// package variable (globallogger.DevMode()) in case the caller has set an env var
	// override, and globallogger.Init will update the global variable.
	currentDevMode := os.Getenv(globallogger.EnvDevelopment) == "true"

	// Initialize sinks
	ss := sinks(append([]Sink{&outputSink{development: currentDevMode}}, s...))
	cores, sinksBuildErr := ss.build()

	// Init the logger first, so that we can log the error if needed, before dealing with
	// sink builder errors
	sync := globallogger.Init(r, currentDevMode, cores)

	if sinksBuildErr != nil {
		// Log the error
		Scoped("log.init").
			Fatal("sinks initialization failed", Error(sinksBuildErr))
	}

	return &PostInitCallbacks{
		Sync:   func() { _ = sync() },
		Update: ss.update,
	}
}
