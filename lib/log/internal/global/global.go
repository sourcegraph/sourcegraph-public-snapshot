package global

import (
	"sync"

	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

var (
	globalLogger     *zap.Logger
	globalLoggerInit sync.Once
)

// Get retrieves the initialized global logger, or panics otherwise.
func Get() *zap.Logger {
	if globalLogger == nil {
		panic("lib/log.Init has not been called")
	}
	return globalLogger
}

// Init initializes the global logger once.
func Init(r otfields.Resource, level zap.AtomicLevel, format encoders.OutputFormat, development bool) {
	globalLoggerInit.Do(func() {
		globalLogger = initLogger(r, level, format, development)
	})
}

func initLogger(r otfields.Resource, level zap.AtomicLevel, format encoders.OutputFormat, development bool) *zap.Logger {
	cfg := zap.Config{
		Level:            level,
		EncoderConfig:    encoders.OpenTelemetryConfig,
		Encoding:         string(format),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},

		// TODO - we collect stacktraces on errors.New, do we need stacktraces on log
		// entries as well?
		DisableStacktrace: true,
	}
	if development {
		cfg.Development = true
		cfg.EncoderConfig = encoders.ApplyDevConfig(cfg.EncoderConfig)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	// Instantiate global
	return logger.With(zap.Object("Resource", &encoders.ResourceEncoder{Resource: r}))
}
