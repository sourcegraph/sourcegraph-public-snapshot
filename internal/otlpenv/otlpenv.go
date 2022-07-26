// Package otlpenv exports getters to read OpenTelemetry protocol configuration options
// based on the official spec: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#configuration-options
package otlpenv

import "os"

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

const (
	// This is a custom default that's also not quite compliant but hopefully close enough (we
	// use 127.0.0.1 instead of localhost, since there's a linter rule banning localhost).
	defaultGRPCCollectorEndpoint     = "http://127.0.0.1:4317"
	defaultHTTPJSONCollectorEndpoint = "http://127.0.0.1:4318"
)

// Endpoint returns the root collector endpoint, NOT per-signal endpoints. We do not yet
// support per-signal endpoints.
//
// See: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#configuration-options
func Endpoint() string {
	return getWithDefault(defaultGRPCCollectorEndpoint,
		"OTEL_EXPORTER_OTLP_ENDPOINT")
}

// HTTPJSONEndpoint returns the configured collector endpoint compatible with
func HTTPJSONEndpoint() string {
	if GetProtocol() == "http/json" {
		// If the default endpoint is the http/json protocol, then just use it directly
		return Endpoint()
	}
	// Otherwsie get a custom value for the http/json protocol endpoint
	return getWithDefault(defaultHTTPJSONCollectorEndpoint,
		"OTEL_EXPORTER_OTLP_HTTP_JSON_ENDPOINT")
}

func GetProtocol() string {
	return getWithDefault("grpc",
		"OTEL_EXPORTER_OTLP_PROTOCOL")
}
