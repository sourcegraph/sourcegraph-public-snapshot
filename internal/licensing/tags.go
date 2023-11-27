package licensing

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/license"
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
	// GPTLLMAccessTag is the license tag that indicates that the licensed instance
	// should be allowed by default to use GPT models in Cody Gateway.
	GPTLLMAccessTag = "gpt"

	// TelemetryEventsExportDisabledTag disables telemery events export EXCEPT
	// for Cody-related events, which we are always allowed to export as part of
	// Cody usage terms: https://sourcegraph.com/terms/cody-notice
	//
	// To completely disable telemetry events export, use PlanAirGappedEnterprise
	TelemetryEventsExportDisabledTag = "disable-telemetry-events-export"
)

// ProductNameWithBrand returns the product name with brand (e.g., "Sourcegraph Enterprise") based
// on the license info.
func ProductNameWithBrand(hasLicense bool, licenseTags []string) string {
	if !hasLicense {
		return "Sourcegraph Free"
	}

	info := &Info{
		Info: license.Info{
			Tags: licenseTags,
		},
	}
	plan := info.Plan()

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
