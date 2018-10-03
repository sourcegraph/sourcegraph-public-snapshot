package licensing

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

// Feature is a product feature that is selectively activated based on the current license key.
type Feature string

// The list of features. For each feature, add a new const here and the checking logic in
// isFeatureEnabled below.
const (
	// FeatureExtensionRegistry is whether publishing extensions to this Sourcegraph instance is
	// allowed. If not, then extensions must be published to Sourcegraph.com. All instances may use
	// extensions published to Sourcegraph.com.
	FeatureExtensionRegistry Feature = "private-extension-registry"
)

// CheckFeature checks whether the feature is activated based on the current license. If it is
// disabled, it returns a non-nil error.
//
// The returned error may implement errcode.PresentationError to indicate that it can be displayed
// directly to the user.
func CheckFeature(feature Feature) error {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("checking feature %q activation", feature))
	}
	if info == nil {
		return errcode.NewPresentationError(fmt.Sprintf("The feature %q is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.", feature))
	}
	if !isFeatureEnabled(*info, feature) {
		return errcode.NewPresentationError(fmt.Sprintf("The feature %q is not activated for Sourcegraph Enterprise Starter. Upgrade to Sourcegraph Enterprise to use this feature.", feature))
	}
	return nil // feature is activated for current license
}

func isFeatureEnabled(info license.Info, feature Feature) bool {
	// Add feature-specific logic here.
	switch feature {
	case FeatureExtensionRegistry:
		// Enterprise Starter does not support a local extension registry.
		return !info.HasTag(EnterpriseStarterTag)
	}
	return false
}
