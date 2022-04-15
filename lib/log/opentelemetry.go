package log

import "go.uber.org/zap/zapcore"

// openTelemetryEncoderConfig configures Zap to comply with the OT logs spec:
// https://opentelemetry.io/docs/reference/specification/logs/data-model/
//
// For what we want output to look like in production, see:
// https://opentelemetry.io/docs/reference/specification/logs/data-model/#example-log-records
var openTelemetryEncoderConfig = zapcore.EncoderConfig{
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
	// Attributes because they are top-level traits, so we just capitalize them and hope
	// for the best.
	CallerKey:     "Caller",
	FunctionKey:   "Function",
	StacktraceKey: "Stacktrace",

	// Defaults
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// Resource represents a service instance.
//
// https://opentelemetry.io/docs/reference/specification/Resource/semantic_conventions/#service
type Resource struct {
	Name      string
	Namespace string
	// InstanceID must be unique for each Name, Namespace pair.
	InstanceID string
	Version    string
}

func (r *Resource) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", r.Name)
	enc.AddString("namespace", r.Namespace)
	enc.AddString("instance.id", r.InstanceID)
	enc.AddString("version", r.Version)
	return nil
}

// attributesNamespace should be included as the last field of all log getters of this
// package.
//
// It logs all fields under 'Attributes' to conform with OpenTelemetry spec
// https://opentelemetry.io/docs/reference/specification/logs/data-model/#field-attributes
var attributesNamespace = Namespace("Attributes")

// https://opentelemetry.io/docs/reference/specification/logs/data-model/#trace-context-fields
type traceContext struct {
	TraceContext
}

func (c *traceContext) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("TraceId", c.TraceID)
	enc.AddString("SpanId", c.SpanID)
	return nil
}
