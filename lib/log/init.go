package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO init stuff still WIP

// Init initializes the log package's global logger as a logger of the given resource.
// It must be called on service startup, NOT on an 'init()' function.
//
// If Init is not called, a dev mode logger is created by Get().
func Init(r Resource, development bool) {
	globalLoggerInit.Do(func() {
		logLevel := zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		go watchLogLevel(logLevel)
		// TODO configurable formats
		globalLogger = initLogger(&r, logLevel, "json", development)
	})
}

// InitForTesting can be used to instantiate the log package for testing.
func InitForTesting(level string) {
	globalLoggerInit.Do(func() {
		globalLogger = initLogger(nil, zap.NewAtomicLevelAt(parseLevel(level)), "console", true)
	})
}

func initLogger(r *Resource, level zap.AtomicLevel, encoding string, development bool) *zap.Logger {
	cfg := zap.Config{
		Level:            level,
		EncoderConfig:    openTelemetryEncoderConfig,
		Encoding:         encoding,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	if development {
		cfg.Development = true
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// Encode time in simpler format, omitting date since in dev this is likely a
		// short-lived instance.
		cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("15:04:05"))
		}
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	// Resource must be configured.
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-resource
	if r == nil {
		r = &Resource{
			Name:      "unknown_service",
			Namespace: "development",
		}
	}

	// Instantiate global
	return logger.With(zap.Object("Resource", r))
}
