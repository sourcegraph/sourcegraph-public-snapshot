package sensitivemetadataallowlist

import (
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
func redactEvent(event *telemetrygatewayv1.Event, mode redactMode, allowedPrivateMetadataKeys []string) {
	// redactNothing (generally in dotcom)
	if mode >= redactNothing {
		return
	}

	// redactMarketingAndUnallowedPrivateMetadataKeys
	event.MarketingTracking = nil
	if mode >= redactMarketingAndUnallowedPrivateMetadataKeys {
		// Do private metadata redaction in this if case, as the next case will strip
		// everything
		for key, value := range event.Parameters.PrivateMetadata.Fields {
			if !slices.Contains(allowedPrivateMetadataKeys, key) {
				// Strip all keys that are NOT in the allowlist
				delete(event.Parameters.PrivateMetadata.Fields, key)
			} else if _, isString := value.Kind.(*structpb.Value_StringValue); !isString {
				// Strip all values that are not strings, even if they ARE in the allowlist
				event.Parameters.PrivateMetadata.Fields[key] =
					structpb.NewStringValue("ERROR: value of allowlisted key was not a string")
			}
		}
		return
	}

	// redactAllSensitive
	if event.Parameters != nil {
		event.Parameters.PrivateMetadata = nil
	}
}
