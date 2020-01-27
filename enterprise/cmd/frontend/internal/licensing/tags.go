package licensing

import "strings"

const (
	// EnterpriseBasicTag is the license tag for Elite.
	EnterpriseBasicTag = "basic"
	// EnterpriseStarterTag is the license tag for Enterprise Starter.
	// Deprecated: This exists for backwards compatibility, future licenses should use EnterprisePlusTags instead.
	EnterpriseStarterTag = "starter"
	// EnterprisePlusTag is the license tag for Enterprise Plus.
	EnterprisePlusTag = "plus"

	// TrueUpUserCountTag is the license tag that indicates that the licensed user count can be
	// exceeded and will be charged later.
	TrueUpUserCountTag = "true-up"
)

var (
	// EnterpriseBasicTags are the license tags for Enterprise.
	EnterpriseBasicTags = []string{EnterpriseBasicTag}
	// EnterpriseStarterTags are the license tags for Enterprise Starter.
	// Deprecated: This exists for backwards compatibility, future licenses should use EnterprisePlusTags instead.
	EnterpriseStarterTags = []string{EnterpriseStarterTag}
	// EnterprisePlusTags are the license tags for Enterprise Plus.
	EnterprisePlusTags = []string{EnterprisePlusTag}
	// EliteTags is the license tags for Elite (intentionally empty because it has no feature restrictions)
	EliteTags = []string{}
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
	if hasTag(EnterpriseBasicTag) {
		name = "Enterprise"
	} else if hasTag(EnterprisePlusTag) {
		name = "Enterprise Plus"
	} else if hasTag(EnterpriseStarterTag) {
		name = "Enterprise Starter"
	} else {
		name = "Elite"
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
