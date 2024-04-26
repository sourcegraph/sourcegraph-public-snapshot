package server

import telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"

// migrateEvents does an in-place migration of any legacy field usages to support
// older exporters.
func migrateEvents(events []*telemetrygatewayv1.Event) {
	for _, ev := range events {
		if ev.Parameters != nil && len(ev.Parameters.LegacyMetadata) > 0 {
			if ev.Parameters.Metadata == nil {
				ev.Parameters.Metadata = make(map[string]float64, len(ev.Parameters.LegacyMetadata))
			}
			for k, v := range ev.Parameters.LegacyMetadata {
				ev.Parameters.Metadata[k] = float64(v)
			}
			ev.Parameters.LegacyMetadata = nil // remove
		}
	}
}
