package sensitivemetadataallowlist

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

var rawAdditionalAllowedEventTypes = env.Get("SRC_TELEMETRY_SENSITIVEMETADATA_ADDITIONAL_ALLOWED_EVENT_TYPES", "",
	"Additional event types to include in sensitivemetadataallowlist.AllowedEventTypes, in comma-separated '${feature}::${action}::${key1}::${key2}' format.")
var additionalAllowedEventTypes = func() []EventType {
	types, err := parseAdditionalAllowedEventTypes(rawAdditionalAllowedEventTypes)
	if err != nil {
		panic(err)
	}
	return types
}()

// AllowedEventTypes denotes a list of all events allowed to export sensitive
// telemetry metadata.
func AllowedEventTypes() EventTypes {
	return eventTypes(append(additionalAllowedEventTypes,
		// Example event for testing.
		EventType{
			Feature: string(telemetry.FeatureExample),
			Action:  string(telemetry.ActionExample),
			AllowedPrivateMetadataKeys: []string{
				"testField",
			},
		},
		// reason for allowlisting
		EventType{
			Feature: "cody.completions",
			Action:  "suggested",
			AllowedPrivateMetadataKeys: []string{
				"languageId",
			},
		},
		EventType{
			Feature: "cody.completions",
			Action:  "accepted",
			AllowedPrivateMetadataKeys: []string{
				"languageId",
			},
		},
	)...)
}

type EventTypes struct {
	types []EventType
	// index of '{feature}.{action}:{allowedfields}' for checking
	index map[string][]string
}

func eventTypes(types ...EventType) EventTypes {
	index := make(map[string][]string, len(types))
	for _, t := range types {
		index[fmt.Sprintf("%s.%s", t.Feature, t.Action)] = t.AllowedPrivateMetadataKeys
	}
	return EventTypes{types: types, index: index}
}

// Redact strips the event of sensitive data based on the allowlist.
//
// 🚨 SECURITY: Be very careful with the redaction modes used here, as it impacts
// what data we export from customer Sourcegraph instances.
func (e EventTypes) Redact(event *telemetrygatewayv1.Event) {
	if dotcom.SourcegraphDotComMode() {
		redactEvent(event, redactNothing, nil)
	} else if keys, allowed := e.IsAllowed(event); allowed {
		redactEvent(event, redactMarketingAndUnallowedPrivateMetadataKeys, keys)
	}
	redactEvent(event, redactAllSensitive, nil)
}

// IsAllowed indicates an event is on the sensitive telemetry allowlist, and the fields that
// are allowed.
func (e EventTypes) IsAllowed(event *telemetrygatewayv1.Event) ([]string, bool) {
	key := fmt.Sprintf("%s.%s", event.GetFeature(), event.GetAction())
	allowedKeys, allowed := e.index[key]
	return allowedKeys, allowed
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
	// AllowedPrivateMetadataKeys is the list of field names permitted to be exported from the `privateMetadata` object.
	AllowedPrivateMetadataKeys []string
}

func (e EventType) validate() error {
	if e.Feature == "" || e.Action == "" {
		return errors.New("feature and action are required")
	}
	if len(e.AllowedPrivateMetadataKeys) == 0 {
		return errors.New("allowedPrivateMetadataKeys are required")
	}
	return nil
}

func init() {
	if err := AllowedEventTypes().validate(); err != nil {
		panic(errors.Wrap(err, "AllowedEvents has invalid event(s)"))
	}
}

func parseAdditionalAllowedEventTypes(config string) ([]EventType, error) {
	if len(config) == 0 {
		return nil, nil
	}

	var types []EventType
	for _, rawType := range strings.Split(config, ",") {
		parts := strings.Split(rawType, "::")
		if len(parts) < 2 {
			return nil, errors.Newf(
				"cannot parse SRC_TELEMETRY_SENSITIVEMETADATA_ADDITIONAL_ALLOWED_EVENT_TYPES value %q",
				rawType)
		}
		// indicates that the user has not specified any allowlisted fields
		if len(parts) < 3 {
			return nil, errors.Newf(
				"cannot parse SRC_TELEMETRY_SENSITIVEMETADATA_ADDITIONAL_ALLOWED_EVENT_TYPES value %q, missing allowlisted fields",
				rawType)
		}
		types = append(types, EventType{
			Feature:                    parts[0],
			Action:                     parts[1],
			AllowedPrivateMetadataKeys: parts[2:],
		})
	}
	return types, nil
}
