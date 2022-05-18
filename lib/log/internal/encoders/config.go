package encoders

import (
	"fmt"

	"github.com/fatih/color"
	"go.uber.org/zap/zapcore"
)

// OpenTelemetryConfig configures Zap to comply with the OT logs spec:
// https://opentelemetry.io/docs/reference/specification/logs/data-model/
//
// For what we want output to look like in production, see:
// https://opentelemetry.io/docs/reference/specification/logs/data-model/#example-log-records
var OpenTelemetryConfig = zapcore.EncoderConfig{
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-instrumentationscope
	NameKey: "InstrumentationScope",
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-timestamp
	TimeKey:    "Timestamp",
	EncodeTime: zapcore.EpochNanosTimeEncoder,
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#severity-fields
	LevelKey:    "SeverityText",
	EncodeLevel: zapcore.CapitalLevelEncoder, // most levels correspond to the OT level text
	// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-body
	MessageKey: "Body",

	// These don't really have an equivalent in the OT spec, and we can't stick it under
	// Attributes because they are top-level traits in Zap, so we just capitalize them and
	// hope for the best.
	CallerKey:     "Caller",
	FunctionKey:   "Function",
	StacktraceKey: "Stacktrace",

	// Defaults
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// applyDevConfig applies options for dev environments to the encoder config
func applyDevConfig(cfg zapcore.EncoderConfig) zapcore.EncoderConfig {
	// Nice colors based on log level
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// Human-readable durations
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	// Make callers clickable in iTerm
	cfg.EncodeCaller = func(entry zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// Link to open the file:line in VS Code.
		url := "vscode://file/" + entry.FullPath()

		// Constructs an escape sequence that iTerm recognizes as a link.
		// See https://iterm2.com/documentation-escape-codes.html
		link := fmt.Sprintf("\x1B]8;;%s\x07%s\x1B]8;;\x07", url, entry.TrimmedPath())

		enc.AppendString(color.New(color.Faint).Sprint(link))
	}
	// Keep output condensed
	cfg.ConsoleSeparator = " "
	// Disabled for now due to verbosity, but we might want to introduce some config for
	// enabling these in the future.
	cfg.FunctionKey = zapcore.OmitKey
	cfg.TimeKey = zapcore.OmitKey
	return cfg
}

func BuildEncoder(format OutputFormat, development bool) (enc zapcore.Encoder) {
	config := OpenTelemetryConfig
	if development {
		config = applyDevConfig(config)
	}

	switch format {
	case OutputConsole:
		return zapcore.NewConsoleEncoder(config)
	case OutputJSON:
		return zapcore.NewJSONEncoder(config)
	default:
		panic("unknown output format")
	}
}
