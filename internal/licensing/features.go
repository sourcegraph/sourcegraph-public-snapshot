package licensing

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Feature interface {
	FeatureName() string
	// Check checks whether the feature is activated on the provided license info.
	// If applicable, it is recommended that Check modifies the feature in-place
	// to reflect the license info (e.g., to set a limit on the number of changesets).
	Check(*Info) error
}

// BasicFeature is a product feature that is selectively activated based on the
// current license key.
type BasicFeature string

func (f BasicFeature) FeatureName() string {
	return string(f)
}

func (f BasicFeature) Check(info *Info) error {
	if info == nil {
		return newFeatureRequiresSubscriptionError(f.FeatureName())
	}

	featureTrimmed := BasicFeature(strings.TrimSpace(f.FeatureName()))

	// Check if the feature is explicitly allowed via license tag.
	hasFeature := func(want Feature) bool {
		for _, t := range info.Tags {
			// We have been issuing licenses with trailing spaces in the tags for a while.
			// Eventually we should be able to remove these `TrimSpace` calls again,
			// as we now guard against that while generating licenses, but there
			// are quite a few "wrong" licenses out there as of today (2021-07-19).
			if BasicFeature(strings.TrimSpace(t)).FeatureName() == want.FeatureName() {
				return true
			}
		}
		return false
	}
	if !(info.Plan().HasFeature(featureTrimmed) || hasFeature(featureTrimmed)) {
		return newFeatureRequiresUpgradeError(f.FeatureName())
	}
	return nil
}

// FeatureBatchChanges is whether Batch Changes on this Sourcegraph instance has been purchased.
type FeatureBatchChanges struct {
	// If true, there is no limit to the number of changesets that can be created.
	Unrestricted bool
	// Maximum number of changesets that can be created per batch change.
	// If Unrestricted is true, this is ignored.
	MaxNumChangesets int
}

func (*FeatureBatchChanges) FeatureName() string {
	return "batch-changes"
}

func (f *FeatureBatchChanges) Check(info *Info) error {
	if info == nil {
		return newFeatureRequiresSubscriptionError(f.FeatureName())
	}

	// If the batch changes tag exists on the license, use unrestricted batch
	// changes.
	if info.HasTag(f.FeatureName()) {
		f.Unrestricted = true
		return nil
	}

	// Otherwise, check the default batch changes feature.
	if info.Plan().HasFeature(f) {
		return nil
	}

	return newFeatureRequiresUpgradeError(f.FeatureName())
}

type FeaturePrivateRepositories struct {
	// If true, there is no limit to the number of private repositories that can be
	// added.
	Unrestricted bool
	// Maximum number of private repositories that can be added. If Unrestricted is
	// true, this is ignored.
	MaxNumPrivateRepos int
}

func (*FeaturePrivateRepositories) FeatureName() string {
	return "private-repositories"
}

func (f *FeaturePrivateRepositories) Check(info *Info) error {
	if info == nil {
		return newFeatureRequiresSubscriptionError(f.FeatureName())
	}

	// If the private repositories tag exists on the license, use unrestricted
	// private repositories.
	if info.HasTag(f.FeatureName()) {
		f.Unrestricted = true
		return nil
	}

	// Otherwise, check the default private repositories feature.
	if info.Plan().HasFeature(f) {
		return nil
	}

	return newFeatureRequiresUpgradeError(f.FeatureName())
}

// Check checks whether the feature is activated based on the current license. If
// it is disabled, it returns a non-nil error.
//
// The returned error may implement errcode.PresentationError to indicate that it
// can be displayed directly to the user. Use IsFeatureNotActivated to
// distinguish between the error reasons.
func Check(feature Feature) error {
	if MockCheckFeature != nil {
		return MockCheckFeature(feature)
	}

	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("checking feature %q activation", feature))
	}

	if !IsLicenseValid() {
		return errors.New("Sourcegraph license is no longer valid")
	}

	return feature.Check(info)
}

// MockCheckFeatureError is for tests that want to mock Check to return a
// specific error or nil (in case of empty string argument).
//
// It returns a cleanup func so callers can use
// `t.Cleanup(licensing.TestingSkipFeatureChecks())` in a test body.
func MockCheckFeatureError(expectedError string) func() {
	MockCheckFeature = func(feature Feature) error {
		if expectedError == "" {
			return nil
		}
		return errors.New(expectedError)
	}
	return func() { MockCheckFeature = nil }
}

// MockCheckFeature is for mocking Check in tests.
var MockCheckFeature func(feature Feature) error

// TestingSkipFeatureChecks is for tests that want to mock Check to always return
// nil (i.e., behave as though the current license enables all features).
//
// It returns a cleanup func so callers can use
// `t.Cleanup(licensing.TestingSkipFeatureChecks())` in a test body.
func TestingSkipFeatureChecks() func() {
	return MockCheckFeatureError("")
}

func NewFeatureNotActivatedError(message string) featureNotActivatedError {
	e := errcode.NewPresentationError(message).(errcode.PresentationError)
	return featureNotActivatedError{e}
}

func newFeatureRequiresSubscriptionError(feature string) featureNotActivatedError {
	msg := fmt.Sprintf("The feature %q is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.", feature)
	return NewFeatureNotActivatedError(msg)
}

func newFeatureRequiresUpgradeError(feature string) featureNotActivatedError {
	msg := fmt.Sprintf("The feature %q is not activated in your Sourcegraph license. Upgrade your Sourcegraph subscription to use this feature.", feature)
	return NewFeatureNotActivatedError(msg)
}

type featureNotActivatedError struct{ errcode.PresentationError }

// IsFeatureNotActivated reports whether err indicates that the license is valid
// but does not activate the feature.
//
// It is used to distinguish between the multiple reasons for errors from Check:
// either failed license verification, or a valid license that does not activate
// a feature (e.g., Enterprise Starter not including an Enterprise-only feature).
func IsFeatureNotActivated(err error) bool {
	// Also check for the pointer type to guard against stupid mistakes.
	return errors.HasType[featureNotActivatedError](err) || errors.HasType[*featureNotActivatedError](err)
}

// IsFeatureEnabledLenient reports whether the current license enables the given
// feature. If there is an error reading the license, it is lenient and returns
// true.
//
// This is useful for callers who don't want to handle errors (usually because
// the user would be prevented from getting to this point if license verification
// had failed, so it's not necessary to handle license verification errors here).
func IsFeatureEnabledLenient(feature Feature) bool {
	return !IsFeatureNotActivated(Check(feature))
}
