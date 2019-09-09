package licensing

import "strings"

const (
	// EnterpriseStarterTag is the license tag for Enterprise Starter (which includes only a subset
	// of Enterprise features).
	EnterpriseStarterTag = "starter"

	// EnterprisePlusTag is the license tag for the Sourcegraph Enterprise Plus tier.
	EnterprisePlusTag = "plus"

	// EliteTag is the license tag for the Sourcegraph Elite tier.
	EliteTag = "elite"

	// TrueUpUserCountTag is the license tag that indicates that the licensed user count can be
	// exceeded and will be charged later.
	TrueUpUserCountTag = "true-up"
)

var (
	// EnterpriseStarterTags is the license tags for Enterprise Starter.
	EnterpriseStarterTags = []string{EnterpriseStarterTag}

	// EnterpriseTags is the license tags for Enterprise (intentionally empty because it has no
	// feature restrictions)
	EnterpriseTags = []string{}

	// EnterprisePlusTags is the license tags for Enterprise Plus.
	EnterprisePlusTags = []string{EnterprisePlusTag}

	// EliteTags is the license tags for Elite.
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

	var name = " Enterprise"
	// DEPRECATED: the Starter product is no longer available for new customers
	if hasTag("starter") {
		name = " Enterprise Starter"
	}

	if hasTag("plus") {
		name = " Enterprise Plus"
	}

	if hasTag("elite") {
		name = " Elite"
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

	return "Sourcegraph" + name
}
