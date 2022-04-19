package otfields

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

// TraceContext represents a trace to associate with log entries.
//
// https://opentelemetry.io/docs/reference/specification/logs/data-model/#trace-context-fields
type TraceContext struct {
	TraceID string
	SpanID  string
}
