package sensitivemetadataallowlist

import (
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

// redactMode dictates how much to redact. The lowest value indicates our
// strictest redaction mode - higher values indicate less redaction.
type redactMode int

const (
	redactAllSensitive redactMode = iota
	// redactMarketing only redacts marketing-related fields.
	redactMarketing
	// redactNothing is only used in dotocm mode.
	redactNothing
)

// 🚨 SECURITY: Be very careful with the redaction mechanisms here, as it impacts
// what data we export from customer Sourcegraph instances.
func redactEvent(event *telemetrygatewayv1.Event, mode redactMode) {
	// redactNothing
	if mode >= redactNothing {
		return
	}

	// redactMarketing
	event.MarketingTracking = nil
	if mode >= redactMarketing {
		return
	}

	// redactAllSensitive
	if event.Parameters != nil {
		event.Parameters.PrivateMetadata = nil
	}
}
