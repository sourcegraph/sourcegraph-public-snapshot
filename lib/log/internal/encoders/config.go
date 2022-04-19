package encoders

import (
	"time"

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
	LevelKey: "SeverityText",
	EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DPanicLevel:
			// DPanicLevel is not really 'FATAL', but we use it as 'Critical' which
			// implies very important, so maybe it works. Alternatively, we can use
			// 'ERROR4' which is allowed in the spec.
			enc.AppendString("FATAL")
		default:
			// All the other levels conform more or less to the OT spec.
			zapcore.CapitalLevelEncoder(l, enc)
		}
	},
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

// ApplyDevConfig applies options for dev environments to the encoder config
func ApplyDevConfig(cfg zapcore.EncoderConfig) zapcore.EncoderConfig {
	// Nice colors based on log level
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// TODO maybe enable this in dev? these get rather verbose, however
	cfg.FunctionKey = zapcore.OmitKey
	// Encode time in simpler format, omitting date since in dev this is likely a
	// short-lived instance.
	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("15:04:05"))
	}
	// Human-readable durations
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	return cfg
}
