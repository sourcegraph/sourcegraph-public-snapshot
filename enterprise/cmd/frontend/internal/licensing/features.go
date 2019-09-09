package licensing

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

// The list of features. For each feature, add a new const here and the checking logic in
// isFeatureEnabled below.
const (
	// FeatureACLs is whether ACLs may be used, such as GitHub or GitLab repository permissions and
	// integration with GitHub/GitLab for user authentication.
	FeatureACLs conf.Feature = "acls"

	// FeatureExtensionRegistry is whether publishing extensions to this Sourcegraph instance is
	// allowed. If not, then extensions must be published to Sourcegraph.com. All instances may use
	// extensions published to Sourcegraph.com.
	FeatureExtensionRegistry conf.Feature = "private-extension-registry"

	// FeatureRemoteExtensionsAllowDisallow is whether the site admin may explictly specify a list
	// of allowed remote extensions and prevent any other remote extensions from being used. It does
	// not apply to locally published extensions.
	FeatureRemoteExtensionsAllowDisallow = "remote-extensions-allow-disallow"
)

func isFeatureEnabled(info license.Info, feature conf.Feature) bool {
	// Allow features to be explicitly enabled/disabled in the license tags.
	if info.HasTag(string(feature)) {
		return true
	}
	if info.HasTag("no-" + string(feature)) {
		return false
	}

	// Add feature-specific logic here.
	switch feature {
	case FeatureACLs:
		// ACLs are technically now only available in Enteprise Plus. But due to existing
		// customers with Enterprise licenses that are using ACLs, only Enterprise Starter
		// is disabled here.
		return !info.HasTag(EnterpriseStarterTag)
	case FeatureExtensionRegistry:
		// Only Sourcegraph Elite supports a local extension registry.
		return info.HasTag(EliteTag)
	case FeatureRemoteExtensionsAllowDisallow:
		// Explictly allowing/disallowing remote extensions by extension ID is technically
		// now only available in Enterprise Plus. But due to existing customers with
		// Enterprise licenses that are using it, only Enterprise Starter is disabled here.
		return !info.HasTag(EnterpriseStarterTag)
	case conf.FeatureGuestUsers:
		// Only Sourcegraph Elite can have unlimited guest users.
		return info.HasTag(EliteTag)
	}
	return false
}

// CheckFeature checks whether the feature is activated based on the current license. If it is
// disabled, it returns a non-nil error.
//
// The returned error may implement errcode.PresentationError to indicate that it can be displayed
// directly to the user. Use IsFeatureNotActivated to distinguish between the error reasons.
func CheckFeature(feature conf.Feature) error {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("checking feature %q activation", feature))
	}
	if info == nil {
		return newFeatureNotActivatedError(fmt.Sprintf("The feature %q is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.", feature))
	}
	if !isFeatureEnabled(*info, feature) {
		return newFeatureNotActivatedError(fmt.Sprintf("The feature %q is not activated for your current license (%s). Upgrade to use this feature.", feature, ProductNameWithBrand(true, info.Tags)))
	}
	return nil // feature is activated for current license
}

func newFeatureNotActivatedError(message string) featureNotActivatedError {
	e := errcode.NewPresentationError(message).(errcode.PresentationError)
	return featureNotActivatedError{e}
}

type featureNotActivatedError struct{ errcode.PresentationError }

// IsFeatureNotActivated reports whether err indicates that the license is valid but does not
// activate the feature.
//
// It is used to distinguish between the multiple reasons for errors from CheckFeature: either
// failed license verification, or a valid license that does not activate a feature (e.g.,
// Enterprise Starter not including an Enterprise-only feature).
func IsFeatureNotActivated(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(featureNotActivatedError)
	if !ok {
		// Also check for the pointer type to guard against stupid mistakes.
		_, ok = err.(*featureNotActivatedError)
	}
	return ok
}

// IsFeatureEnabledLenient reports whether the current license enables the given feature. If there
// is an error reading the license, it is lenient and returns true.
//
// This is useful for callers who don't want to handle errors (usually because the user would be
// prevented from getting to this point if license verification had failed, so it's not necessary to
// handle license verification errors here).
func IsFeatureEnabledLenient(feature conf.Feature) bool {
	return !IsFeatureNotActivated(CheckFeature(feature))
}
