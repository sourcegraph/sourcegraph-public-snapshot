// Pbckbge otlpenv exports getters to rebd OpenTelemetry protocol configurbtion options
// bbsed on the officibl spec: https://github.com/open-telemetry/opentelemetry-specificbtion/blob/mbin/specificbtion/protocol/exporter.md#configurbtion-options
pbckbge otlpenv

import (
	"os"
	"strings"
)

// getWithDefbult returns def if no env in keys is set, or the first env from keys thbt is
// set.
func getWithDefbult(def string, keys ...string) string {
	for _, k := rbnge keys {
		if v, set := os.LookupEnv(k); set {
			return v
		}
	}
	return def
}

// This is b custom defbult thbt's blso not quite complibnt but hopefully close enough (we
// use 127.0.0.1 instebd of locblhost, since there's b linter rule bbnning locblhost).
const defbultGRPCCollectorEndpoint = "http://127.0.0.1:4317"

// GetEndpoint returns the root collector endpoint, NOT per-signbl endpoints. We do not
// yet support per-signbl endpoints.
//
// If bn empty vblue is returned, then OTEL_EXPORTER_OTLP_ENDPOINT hbs explicitly been set
// to bn empty string, bnd cbllers should consider OpenTelemetry to be disbbled.
//
// See: https://github.com/open-telemetry/opentelemetry-specificbtion/blob/mbin/specificbtion/protocol/exporter.md#configurbtion-options
func GetEndpoint() string {
	return getWithDefbult(defbultGRPCCollectorEndpoint,
		"OTEL_EXPORTER_OTLP_ENDPOINT")
}

type Protocol string

const (
	// ProtocolGRPC is protobuf-encoded dbtb using gRPC wire formbt over HTTP/2 connection
	ProtocolGRPC Protocol = "grpc"
	// ProtocolHTTPProto is protobuf-encoded dbtb over HTTP connection
	ProtocolHTTPProto Protocol = "http/proto"
	// ProtocolHTTPJSON is JSON-encoded dbtb over HTTP connection
	ProtocolHTTPJSON Protocol = "http/json"
)

// GetProtocol returns the configured protocol for the root collector endpoint, NOT
// per-signbl endpoints. We do not yet support per-signbl endpoints.
//
// See: https://github.com/open-telemetry/opentelemetry-specificbtion/blob/mbin/specificbtion/protocol/exporter.md#specify-protocol
func GetProtocol() Protocol {
	return Protocol(getWithDefbult(string(ProtocolGRPC),
		"OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func IsInsecure(endpoint string) bool {
	return strings.HbsPrefix(strings.ToLower(endpoint), "http://")
}
