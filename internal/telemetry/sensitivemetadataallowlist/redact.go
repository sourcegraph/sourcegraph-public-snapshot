package sensitivemetadataallowlist

import (
	"fmt"
	"slices"

	"google.golang.org/protobuf/types/known/structpb"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

// redactMode dictates how much to redact. The lowest value indicates our
// strictest redaction mode - higher values indicate less redaction.
type redactMode int

const (
	redactAllSensitive redactMode = iota
	// redactMarketingAndUnallowedPrivateMetadataKeys only redacts marketing-related fields as well
	// as unallowed private metadata keys.
	redactMarketingAndUnallowedPrivateMetadataKeys
	// redactNothing is only used in dotocm mode.
	redactNothing
)

// 🚨 SECURITY: Be very careful with the redaction mechanisms here, as it impacts
// what data we export from customer Sourcegraph instances.
func redactEvent(event *telemetrygatewayv1.Event, mode redactMode, allowedPrivateMetadataKeys []string) redactMode {
	// redactNothing (generally only in dotcom)
	if mode >= redactNothing {
		return mode
	}

	// redactMarketingAndUnallowedPrivateMetadataKeys
	event.MarketingTracking = nil
	if mode >= redactMarketingAndUnallowedPrivateMetadataKeys {
		// Do private metadata redaction in this if case, as the next case will strip
		// everything
		if event.Parameters == nil || event.Parameters.PrivateMetadata == nil {
			return mode
		}
		for key, value := range event.GetParameters().GetPrivateMetadata().GetFields() {
			if !slices.Contains(allowedPrivateMetadataKeys, key) {
				// Strip all keys that are NOT in the allowlist, as they are considered sensitive.
				// No need to check data types, since we're only dealing with keys.
				delete(event.Parameters.PrivateMetadata.Fields, key)
			} else if _, isString := value.Kind.(*structpb.Value_StringValue); !isString {
				// Strip all non-string values,even if they are in the allowlist
				// 🚨 This prevents exporting arbitrary data types with deep values, which could lead to uncontrolled data exposure.
				event.Parameters.PrivateMetadata.Fields[key] =
					structpb.NewStringValue(fmt.Sprintf("ERROR: value of allowlisted key was not a string, got: %T", value.Kind))
			}
		}
		return mode
	}

	// redactAllSensitive
	if event.Parameters != nil {
		event.Parameters.PrivateMetadata = nil
	}
	return mode
}
