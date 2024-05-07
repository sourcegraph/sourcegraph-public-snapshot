// Package v1 publishes telemetrygateway V1 bindings for both internal
// (single-tenant Sourcegraph) and external (standalone Sourcegraph-managed
// services) consumption. This package also includes standard defaults and
// helpers for Telemetry Gateway integrations.
package v1

import (
	"github.com/google/uuid"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DefaultEventIDFunc is the default generator for telemetry event IDs.
// We currently use V7, which is time-ordered, making them useful for event IDs.
// https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format-03#name-uuid-version-7
var DefaultEventIDFunc = func() string {
	return uuid.Must(uuid.NewV7()).String()
}

// featureActionRegex is used to validate feature and action names. Values must:
// - Start with a lowercase letter
// - Contain only letters, and dashes and dots as delimters
// - Not contain any whitespace
var featureActionRegex = regexp.MustCompile(`^[a-z][a-zA-Z-\.]+$`)

// featureActionMaxLength is the maximum length of a feature or action name.
const featureActionMaxLength = 64

// ValidateEventFeatureAction validates the given feature and action names. It
// should be used where event features and actions are provided by a client.
func ValidateEventFeatureAction(feature, action string) error {
	if feature == "" || action == "" {
		return errors.New("'feature', 'action' must both be provided")
	}
	if len(feature) > featureActionMaxLength {
		return errors.New("'feature' must be less than 64 characters")
	}
	if len(action) > featureActionMaxLength {
		return errors.New("'action' must be less than 64 characters")
	}
	if !featureActionRegex.MatchString(feature) {
		return errors.New("'feature' must start with a lowercase letter and contain only letters, dashes, and dots")
	}
	if !featureActionRegex.MatchString(action) {
		return errors.New("'action' must start with a lowercase letter and contain only letters, dashes, and dots")
	}
	return nil
}
