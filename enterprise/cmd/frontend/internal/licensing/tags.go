package licensing

import "strings"

const (
	// TrueUpUserCountTag is the license tag that indicates that the licensed user count can be
	// exceeded and will be charged later.
	TrueUpUserCountTag = "true-up"
)

// ProductNameWithBrand returns the product name with brand (e.g., "Sourcegraph Enterprise") based
// on the license info.
func ProductNameWithBrand(hasLicense bool, licenseTags []string) string {
	if !hasLicense {
		return "Sourcegraph Free"
	}

	hasTag := func(tag string) bool {
		for _, t := range licenseTags {
			if tag == t {
				return true
			}
		}
		return false
	}

	baseName := "Sourcegraph Enterprise"
	var name string

	if hasTag("team") {
		baseName = "Sourcegraph Team"
	} else if hasTag("starter") {
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

	return baseName + name
}
