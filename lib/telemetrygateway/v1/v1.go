// Package v1 publishes telemetrygateway V1 bindings for external consumption
// from standalone Sourcegraph managed services, and also includes some basic
// helpers.
//
// The source for the API specification lives in
// 'internal/telemetrygateway/v1/telemetrygateway.proto'.
package v1

import "github.com/google/uuid"

// DefaultEventIDFunc is the default generator for telemetry event IDs.
// We currently use V7, which is time-ordered, making them useful for event IDs.
// https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format-03#name-uuid-version-7
var DefaultEventIDFunc = func() string {
	return uuid.Must(uuid.NewV7()).String()
}
