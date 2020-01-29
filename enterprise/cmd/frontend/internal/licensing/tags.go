package licensing

import "strings"

// License tags for different tiers and options.
const (
	// Deprecated: This exists for backwards compatibility, future licenses should use EnterprisePlusTag instead.
	EnterpriseStarterTag = "starter"
	// EnterprisePlusTag is the license tag for Enterprise Plus.
	EnterprisePlusTag = "plus"
	// EliteTag is the license tag for Elite.
	EliteTag = "elite"

	// TrueUpUserCountTag is the license tag that indicates that the licensed user count can be
	// exceeded and will be charged later.
	TrueUpUserCountTag = "true-up"
)

var (
	// EnterpriseTags is the license tags for Enterprise.
	// For historical reason, we have no license tags for the Enterprise tier.
	EnterpriseTags []string
	// EnterpriseStarterTags are the license tags for Enterprise Starter.
	// Deprecated: This exists for backwards compatibility, future licenses should use EnterprisePlusTags instead.
	EnterpriseStarterTags = []string{EnterpriseStarterTag}
	// EnterprisePlusTags are the license tags for Enterprise Plus.
	EnterprisePlusTags = []string{EnterprisePlusTag}
	// EliteTag are the license tags for Elite.
	EliteTags = []string{EliteTag}
)

// ProductNameWithBrand returns the product name with brand (e.g., "Sourcegraph Enterprise") based
// on the license info.
func ProductNameWithBrand(hasLicense bool, licenseTags []string) string {
	if !hasLicense {
		return "Sourcegraph Core"
	}

	hasTag := func(tag string) bool {
		for _, t := range licenseTags {
			if tag == t {
				return true
			}
		}
		return false
	}

	var name string
	if hasTag(EliteTag) {
		name = "Elite"
	} else if hasTag(EnterprisePlusTag) {
		name = "Enterprise Plus"
	} else if hasTag(EnterpriseStarterTag) {
		name = "Enterprise Starter"
	} else {
		name = "Enterprise"
	}

	var misc []string
	if hasTag("trial") {
		misc = append(misc, "trial")
	}
	if hasTag("dev") {
		misc = append(misc, "dev use only")
	}
	if len(misc) > 0 {
		name += " (" + strings.Join(misc, ", ") + ")"
	}

	return "Sourcegraph " + name
}
