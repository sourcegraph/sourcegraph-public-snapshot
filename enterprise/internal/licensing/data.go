package licensing

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

	// PlanFree0 is the default plan if no license key is set before 4.5.
	PlanFree0 Plan = "free-0"

	// PlanFree1 is the default plan if no license key is set from 4.5 onwards.
	PlanFree1 Plan = "free-1"

	// PlanAirGappedEnterprise is the same PlanEnterprise1 but with FeatureAllowAirGapped, and works starting from 5.1.
	PlanAirGappedEnterprise Plan = "enterprise-air-gap-0"
)

var AllPlans = []Plan{
	PlanOldEnterpriseStarter,
	PlanOldEnterprise,
	PlanTeam0,
	PlanEnterprise0,

	PlanBusiness0,
	PlanEnterprise1,
	PlanEnterpriseExtension,
	PlanFree0,
	PlanFree1,
	PlanAirGappedEnterprise,
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

	// FeatureSCIM is whether SCIM User Management has been purchased on this instance.
	FeatureSCIM BasicFeature = "SCIM"

	// FeatureCody is whether or not Cody and embeddings has been purchased on this instance.
	FeatureCody BasicFeature = "cody"

	// FeatureAllowAirGapped is whether or not air gapped mode is allowed on this instance.
	FeatureAllowAirGapped BasicFeature = "allow-air-gapped"
)

var AllFeatures = []Feature{
	FeatureSSO,
	FeatureACLs,
	FeatureExplicitPermissionsAPI,
	FeatureExtensionRegistry,
	FeatureRemoteExtensionsAllowDisallow,
	FeatureBranding,
	FeatureCampaigns,
	FeatureMonitoring,
	FeatureBackupAndRestore,
	FeatureCodeInsights,
	&FeatureBatchChanges{},
	FeatureSCIM,
	FeatureAllowAirGapped,
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
			&FeaturePrivateRepositories{Unrestricted: true},
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
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureMonitoring,
			FeatureBackupAndRestore,
			FeatureCodeInsights,
			FeatureSCIM,
			FeatureCody,
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
			&FeaturePrivateRepositories{Unrestricted: true},
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
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureSCIM,
			FeatureCody,
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
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureCodeInsights,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
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
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
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
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
	PlanFree0: {
		Features: []Feature{
			FeatureSSO,
			FeatureMonitoring,
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{Unrestricted: true},
		},
		ExpiredFeatures: []Feature{
			FeatureSSO,
			FeatureMonitoring,
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{Unrestricted: true},
		},
	},
	PlanFree1: {
		Features: []Feature{
			FeatureMonitoring,
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{MaxNumPrivateRepos: 1},
		},
		ExpiredFeatures: []Feature{
			FeatureMonitoring,
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{MaxNumPrivateRepos: 1},
		},
	},
	PlanAirGappedEnterprise: {
		Features: []Feature{
			FeatureACLs,
			FeatureCampaigns,
			FeatureCodeInsights,
			&FeatureBatchChanges{Unrestricted: true},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
			FeatureAllowAirGapped,
		},
		ExpiredFeatures: []Feature{
			FeatureACLs,
			FeatureSSO,
		},
	},
}
