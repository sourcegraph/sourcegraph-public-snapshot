package licensing

import (
	"strings"
)

const (
	// EnterpriseStarterTag is the license tag for Enterprise Starter (which includes only a subset
	// of Enterprise features).
	EnterpriseStarterTag = "starter"

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
	if hasTag("starter") {
		name = " Starter"
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

	return "Sourcegraph Enterprise" + name
}
