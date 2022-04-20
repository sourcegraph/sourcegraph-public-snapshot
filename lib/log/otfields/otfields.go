package otfields

import "go.uber.org/zap"

// Resource represents a service instance.
//
// https://opentelemetry.io/docs/reference/specification/Resource/semantic_conventions/#service
type Resource struct {
	// Name is the logical name of the service. Must be the same for all instances of
	// horizontally scaled services. Optional, and falls back to 'unknown_service' as per
	// the OpenTelemetry spec.
	Name string
	// Namespace helps to distinguish a group of services, for example the team name that
	// owns a group of services. Optional.
	Namespace string
	// Version is the version string of the service API or implementation. For Sourcegraph
	// services, this should be from 'internal/version.Version()'
	Version string
}

// TraceContext represents a trace to associate with log entries.
//
// https://opentelemetry.io/docs/reference/specification/logs/data-model/#trace-context-fields
type TraceContext struct {
	TraceID string
	SpanID  string
}

// attributesNamespace is the namespace under which all arbitrary fields are logged, as
// per the OpenTelemetry spec.
//
// Only for internal use.
var AttributesNamespace = zap.Namespace("Attributes")
