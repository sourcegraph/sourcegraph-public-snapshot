package sensitivemetadataallowlist

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

// AllowedEventTypes denotes a list of all events allowed to export sensitive
// telemetry metadata.
func AllowedEventTypes() EventTypes {
	return eventTypes(
		// Example event for testing.
		EventType{
			Feature: string(telemetry.FeatureExample),
			Action:  string(telemetry.ActionExample),
		},
	)
}

type EventTypes struct {
	types []EventType
	// index of '{feature}.{action}' for checking
	index map[string]struct{}
}

func eventTypes(types ...EventType) EventTypes {
	index := make(map[string]struct{}, len(types))
	for _, t := range types {
		index[fmt.Sprintf("%s.%s", t.Feature, t.Action)] = struct{}{}
	}
	return EventTypes{types: types, index: index}
}

// Redact strips the event of sensitive data based on the allowlist.
//
// 🚨 SECURITY: Be very careful with the redaction modes used here, as it impacts
// what data we export from customer Sourcegraph instances.
func (e EventTypes) Redact(event *telemetrygatewayv1.Event) {
	rm := redactAllSensitive
	if envvar.SourcegraphDotComMode() {
		rm = redactNothing
	} else if e.IsAllowed(event) {
		rm = redactMarketing
	}
	redactEvent(event, rm)
}

// IsAllowed indicates an event is on the sensitive telemetry allowlist.
func (e EventTypes) IsAllowed(event *telemetrygatewayv1.Event) bool {
	key := fmt.Sprintf("%s.%s", event.GetFeature(), event.GetAction())
	_, allowed := e.index[key]
	return allowed
}

func (e EventTypes) validate() error {
	for _, t := range e.types {
		if err := t.validate(); err != nil {
			return err
		}
	}
	return nil
}

type EventType struct {
	Feature string
	Action  string

	// Future: maybe restrict to specific, known private metadata fields as well
}

func (e EventType) validate() error {
	if e.Feature == "" || e.Action == "" {
		return errors.New("feature and action are required")
	}
	return nil
}

func init() {
	if err := AllowedEventTypes().validate(); err != nil {
		panic(errors.Wrap(err, "AllowedEvents has invalid event(s)"))
	}
}
