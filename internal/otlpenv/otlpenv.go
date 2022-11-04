// Package otlpenv exports getters to read OpenTelemetry protocol configuration options
// based on the official spec: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#configuration-options
package otlpenv

import (
	"os"
	"strings"
)

// getWithDefault returns def if no env in keys is set, or the first env from keys that is
// set.
func getWithDefault(def string, keys ...string) string {
	for _, k := range keys {
		if v, set := os.LookupEnv(k); set {
			return v
		}
	}
	return def
}

// This is a custom default that's also not quite compliant but hopefully close enough (we
// use 127.0.0.1 instead of localhost, since there's a linter rule banning localhost).
const defaultGRPCCollectorEndpoint = "http://127.0.0.1:4317"

// GetEndpoint returns the root collector endpoint, NOT per-signal endpoints. We do not
// yet support per-signal endpoints.
//
// If an empty value is returned, then OTEL_EXPORTER_OTLP_ENDPOINT has explicitly been set
// to an empty string, and callers should consider OpenTelemetry to be disabled.
//
// See: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#configuration-options
func GetEndpoint() string {
	return getWithDefault(defaultGRPCCollectorEndpoint,
		"OTEL_EXPORTER_OTLP_ENDPOINT")
}

type Protocol string

const (
	// ProtocolGRPC is protobuf-encoded data using gRPC wire format over HTTP/2 connection
	ProtocolGRPC Protocol = "grpc"
	// ProtocolHTTPProto is protobuf-encoded data over HTTP connection
	ProtocolHTTPProto Protocol = "http/proto"
	// ProtocolHTTPJSON is JSON-encoded data over HTTP connection
	ProtocolHTTPJSON Protocol = "http/json"
)

// GetProtocol returns the configured protocol for the root collector endpoint, NOT
// per-signal endpoints. We do not yet support per-signal endpoints.
//
// See: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#specify-protocol
func GetProtocol() Protocol {
	return Protocol(getWithDefault(string(ProtocolGRPC),
		"OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func IsInsecure(endpoint string) bool {
	return strings.HasPrefix(strings.ToLower(endpoint), "http://")
}
