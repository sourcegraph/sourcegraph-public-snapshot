package globallogger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

var (
	devMode          bool
	globalLogger     *zap.Logger
	globalLoggerInit sync.Once
)

func DevMode() bool { return devMode }

// Get retrieves the initialized global logger, or panics otherwise (unless safe is true,
// in which case a no-op logger is returned)
func Get(safe bool) *zap.Logger {
	if !IsInitialized() {
		if safe {
			return zap.NewNop()
		} else {
			panic("global logger not initialized - have you called lib/log.Init or lib/logtest.Init?")
		}
	}
	return globalLogger
}

// Init initializes the global logger once. Subsequent calls are no-op. Returns the
// callback to sync the root core.
func Init(r otfields.Resource, level zapcore.LevelEnabler, format encoders.OutputFormat, development bool) func() error {
	globalLoggerInit.Do(func() {
		globalLogger = initLogger(r, level, format, development)
	})
	return globalLogger.Sync
}

// IsInitialized indicates if the global logger is initialized.
func IsInitialized() bool {
	return globalLogger != nil
}

func initLogger(r otfields.Resource, level zapcore.LevelEnabler, format encoders.OutputFormat, development bool) *zap.Logger {
	// Set global
	devMode = development

	logSink, errSink, err := openStderrSinks()
	if err != nil {
		panic(err.Error())
	}

	options := []zap.Option{zap.ErrorOutput(errSink), zap.AddCaller()}
	if development {
		options = append(options, zap.Development())
	}

	logger := zap.New(zapcore.NewCore(
		encoders.BuildEncoder(format, development),
		logSink,
		level,
	), options...)

	if development {
		return logger
	}

	// If not in development, log OpenTelemetry Resource field and generate an InstanceID
	// to uniquely identify this resource.
	//
	// See examples: https://opentelemetry.io/docs/reference/specification/logs/data-model/#example-log-records
	if r.InstanceID == "" {
		r.InstanceID = uuid.New().String()
	}
	return logger.With(zap.Object("Resource", &encoders.ResourceEncoder{Resource: r}))
}

// copied from https://sourcegraph.com/github.com/uber-go/zap/-/blob/config.go?L249
func openStderrSinks() (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	sink, closeOut, err := zap.Open("stderr")
	if err != nil {
		return nil, nil, err
	}
	errSink, _, err := zap.Open("stderr")
	if err != nil {
		closeOut()
		return nil, nil, err
	}
	return sink, errSink, nil
}
