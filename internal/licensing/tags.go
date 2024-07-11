package licensing

import (
	"strings"
)

const (
	// TrialTag denotes trial licenses.
	TrialTag = "trial"
	// TrueUpUserCountTag is the license tag that indicates that the licensed user count can be
	// exceeded and will be charged later.
	TrueUpUserCountTag = "true-up"
	// InternalTag denotes Sourcegraph-internal tags
	InternalTag = "internal"
	// DevTag denotes licenses used in development environments
	DevTag = "dev"

	// TelemetryEventsExportDisabledTag disables telemery events export EXCEPT
	// for Cody-related events, which we are always allowed to export as part of
	// Cody usage terms: https://sourcegraph.com/terms/cody-notice
	//
	// To completely disable telemetry events export, use FeatureAllowAirGapped.
	TelemetryEventsExportDisabledTag = "disable-telemetry-events-export"
)

// ProductNameWithBrand returns the product name with brand (e.g., "Sourcegraph Enterprise") based
// on the license info.
func ProductNameWithBrand(licenseTags []string) string {
	plan := PlanFromTags(licenseTags)

	details, ok := planDetails[plan]
	if !ok {
		return "Unrecognized plan"
	}

	name := details.DisplayName

	hasTag := func(tag string) bool {
		for _, t := range licenseTags {
			if tag == t {
				return true
			}
		}
		return false
	}

	var misc []string
	if hasTag(TrialTag) {
		misc = append(misc, "trial")
	}
	if hasTag(DevTag) {
		misc = append(misc, "dev use only")
	}
	if hasTag(InternalTag) {
		misc = append(misc, "internal use only")
	}
	if len(misc) > 0 {
		name += " (" + strings.Join(misc, ", ") + ")"
	}

	return name
}
