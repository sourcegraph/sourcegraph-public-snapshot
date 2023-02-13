package licensing

import "fmt"

// The list of plans.
const (
	// PlanOldEnterpriseStarter is the old "Enterprise Starter" plan.
	PlanOldEnterpriseStarter Plan = "old-starter-0"
	// PlanOldEnterprise is the old "Enterprise" plan.
	PlanOldEnterprise Plan = "old-enterprise-0"

	// PlanTeam0 is the "Team" plan pre-4.0.
	PlanTeam0 Plan = "team-0"
	// PlanEnterprise0 is the "Enterprise" plan pre-4.0.
	PlanEnterprise0 Plan = "enterprise-0"

	// PlanBusiness0 is the "Business" plan for 4.0.
	PlanBusiness0 Plan = "business-0"
	// PlanEnterprise1 is the "Enterprise" plan for 4.0.
	PlanEnterprise1 Plan = "enterprise-1"

	// PlanEnterpriseExtension is for customers who require an extended trial on a new Sourcegraph 4.4.2 instance.
	PlanEnterpriseExtension Plan = "enterprise-extension"

	// PlanFree0 is the default plan if no license key is set.
	PlanFree0 Plan = "free-0"
)

var allPlans = []Plan{
	PlanOldEnterpriseStarter,
	PlanOldEnterprise,
	PlanTeam0,
	PlanEnterprise0,

	PlanBusiness0,
	PlanEnterprise1,
	PlanEnterpriseExtension,
	PlanFree0,
}

// The list of features. For each feature, add a new const here and the checking logic in
// isFeatureEnabled.
const (
	// FeatureSSO is whether non-builtin authentication may be used, such as GitHub
	// OAuth, GitLab OAuth, SAML, and OpenID.
	FeatureSSO BasicFeature = "sso"

	// FeatureACLs is whether the Background Permissions Syncing may be be used for
	// setting repository permissions.
	FeatureACLs BasicFeature = "acls"

	// FeatureExplicitPermissionsAPI is whether the Explicit Permissions API may be be used for
	// setting repository permissions.
	FeatureExplicitPermissionsAPI BasicFeature = "explicit-permissions-api"

	// FeatureExtensionRegistry is whether publishing extensions to this Sourcegraph instance has been
	// purchased. If not, then extensions must be published to Sourcegraph.com. All instances may use
	// extensions published to Sourcegraph.com.
	FeatureExtensionRegistry BasicFeature = "private-extension-registry"

	// FeatureRemoteExtensionsAllowDisallow is whether explicitly specify a list of allowed remote
	// extensions and prevent any other remote extensions from being used has been purchased. It
	// does not apply to locally published extensions.
	FeatureRemoteExtensionsAllowDisallow BasicFeature = "remote-extensions-allow-disallow"

	// FeatureBranding is whether custom branding of this Sourcegraph instance has been purchased.
	FeatureBranding BasicFeature = "branding"

	// FeatureCampaigns is whether campaigns (now: batch changes) on this Sourcegraph instance has been purchased.
	//
	// DEPRECATED: See FeatureBatchChanges.
	FeatureCampaigns BasicFeature = "campaigns"

	// FeatureMonitoring is whether monitoring on this Sourcegraph instance has been purchased.
	FeatureMonitoring BasicFeature = "monitoring"

	// FeatureBackupAndRestore is whether builtin backup and restore on this Sourcegraph instance
	// has been purchased.
	FeatureBackupAndRestore BasicFeature = "backup-and-restore"

	// FeatureCodeInsights is whether Code Insights on this Sourcegraph instance has been purchased.
	FeatureCodeInsights BasicFeature = "code-insights"
)

// FeatureBatchChanges is whether Batch Changes on this Sourcegraph instance has been purchased.
type FeatureBatchChanges struct {
	// If true, there is no limit to the number of changesets that can be created.
	Unrestricted bool
	// Maximum number of changesets that can be created per batch change. If Unrestricted is true, this is ignored.
	MaxNumChangesets int
}

func (*FeatureBatchChanges) FeatureName() string {
	return "batch-changes"
}

func (f *FeatureBatchChanges) Check(info *Info) error {
	if info == nil {
		return NewFeatureNotActivatedError(fmt.Sprintf("The feature %q is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.", f.FeatureName()))
	}

	// If the deprecated campaigns are enabled, use unrestricted batch changes
	if FeatureCampaigns.Check(info) == nil && !info.IsExpired() {
		f.Unrestricted = true
		return nil
	}

	// If the batch changes tag exists on the license, use unrestricted batch changes
	if info.HasTag(f.FeatureName()) && !info.IsExpired() {
		f.Unrestricted = true
		return nil
	}

	// Otherwise, check the default batch changes feature
	if info.Plan().HasFeature(f, info.IsExpired()) {
		return nil
	}

	return NewFeatureNotActivatedError(fmt.Sprintf("The feature %q is not activated in your Sourcegraph license. Upgrade your Sourcegraph subscription to use this feature.", f.FeatureName()))
}

type PlanDetails struct {
	Features []Feature
	// ExpiredFeatures are the features that still work after the plan is expired.
	ExpiredFeatures []Feature
}

// planDetails defines the features that are enabled for each plan.
var planDetails = map[Plan]PlanDetails{
	PlanOldEnterpriseStarter: {
		Features: []Feature{
			&FeatureBatchChanges{MaxNumChangesets: 10},
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
	PlanOldEnterprise: {
		Features: []Feature{
			FeatureSSO,
			FeatureACLs,
			FeatureExplicitPermissionsAPI,
			FeatureExtensionRegistry,
			FeatureRemoteExtensionsAllowDisallow,
			FeatureBranding,
			FeatureCampaigns,
			&FeatureBatchChanges{Unrestricted: true},
			FeatureMonitoring,
			FeatureBackupAndRestore,
			FeatureCodeInsights,
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
	PlanTeam0: {
		Features: []Feature{
			FeatureACLs,
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			&FeatureBatchChanges{MaxNumChangesets: 10},
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
	PlanEnterprise0: {
		Features: []Feature{
			FeatureACLs,
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			&FeatureBatchChanges{MaxNumChangesets: 10},
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},

	PlanBusiness0: {
		Features: []Feature{
			FeatureACLs,
			FeatureCampaigns,
			&FeatureBatchChanges{Unrestricted: true},
			FeatureCodeInsights,
			FeatureSSO,
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
	PlanEnterprise1: {
		Features: []Feature{
			FeatureACLs,
			FeatureCampaigns,
			FeatureCodeInsights,
			&FeatureBatchChanges{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
	PlanEnterpriseExtension: {
		Features: []Feature{
			FeatureACLs,
			FeatureCampaigns,
			FeatureCodeInsights,
			&FeatureBatchChanges{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
	PlanFree0: {Features: []Feature{
		FeatureSSO,
		FeatureMonitoring,
		&FeatureBatchChanges{MaxNumChangesets: 10},
	}},
}
