// Package v1 publishes telemetrygateway V1 bindings for both internal
// (single-tenant Sourcegraph) and external (standalone Sourcegraph-managed
// services) consumption. This package also includes standard defaults and
// helpers for Telemetry Gateway integrations.
package v1

import "github.com/google/uuid"

// DefaultEventIDFunc is the default generator for telemetry event IDs.
// We currently use V7, which is time-ordered, making them useful for event IDs.
// https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format-03#name-uuid-version-7
var DefaultEventIDFunc = func() string {
	return uuid.Must(uuid.NewV7()).String()
}
