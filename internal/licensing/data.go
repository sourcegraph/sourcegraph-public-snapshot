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

	// FeatureCodeInsights is whether Code Insights on this Sourcegraph instance has been purchased.
	FeatureCodeInsights BasicFeature = "code-insights"

	// FeatureSCIM is whether SCIM User Management has been purchased on this instance.
	FeatureSCIM BasicFeature = "SCIM"

	// FeatureCody is whether or not Cody and embeddings has been purchased on this instance.
	FeatureCody BasicFeature = "cody"

	// FeatureAllowAirGapped is whether or not air gapped mode is allowed on this instance.
	FeatureAllowAirGapped BasicFeature = "allow-air-gapped"

	// FeatureCodeMonitors is whether code monitors is allowed on this Sourcegraph instance.
	FeatureCodeMonitors BasicFeature = "code-monitors"

	// FeatureNotebooks is whether the notebooks feature is allowed on this Sourcegraph instance.
	FeatureNotebooks BasicFeature = "notebooks"

	// FeatureCodeSearch is whether the code search feature suite is allowed on this Sourcegraph instance.
	FeatureCodeSearch BasicFeature = "code-search"
)

var AllFeatures = []Feature{
	FeatureSSO,
	FeatureACLs,
	FeatureExplicitPermissionsAPI,
	FeatureCodeInsights,
	&FeatureBatchChanges{},
	FeatureSCIM,
	FeatureAllowAirGapped,
	FeatureCodeMonitors,
	FeatureNotebooks,
	FeatureCodeSearch,
}

type PlanDetails struct {
	Features []Feature
}

// planDetails defines the features that are enabled for each plan.
var planDetails = map[Plan]PlanDetails{
	PlanOldEnterpriseStarter: {
		Features: []Feature{
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
	PlanOldEnterprise: {
		Features: []Feature{
			FeatureSSO,
			FeatureACLs,
			FeatureExplicitPermissionsAPI,
			&FeatureBatchChanges{Unrestricted: true},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureCodeInsights,
			FeatureSCIM,
			FeatureCody,
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
	PlanTeam0: {
		Features: []Feature{
			FeatureACLs,
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
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
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},

	PlanBusiness0: {
		Features: []Feature{
			FeatureACLs,
			&FeatureBatchChanges{Unrestricted: true},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureCodeInsights,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
	PlanEnterprise1: {
		Features: []Feature{
			FeatureACLs,
			FeatureCodeInsights,
			&FeatureBatchChanges{Unrestricted: true},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
	PlanEnterpriseExtension: {
		Features: []Feature{
			FeatureACLs,
			FeatureCodeInsights,
			&FeatureBatchChanges{Unrestricted: true},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
	PlanFree0: {
		Features: []Feature{
			FeatureSSO,
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
	PlanFree1: {
		Features: []Feature{
			&FeatureBatchChanges{MaxNumChangesets: 10},
			&FeaturePrivateRepositories{MaxNumPrivateRepos: 1},
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
	PlanAirGappedEnterprise: {
		Features: []Feature{
			FeatureACLs,
			FeatureCodeInsights,
			&FeatureBatchChanges{Unrestricted: true},
			&FeaturePrivateRepositories{Unrestricted: true},
			FeatureExplicitPermissionsAPI,
			FeatureSSO,
			FeatureSCIM,
			FeatureCody,
			FeatureAllowAirGapped,
			FeatureCodeMonitors,
			FeatureNotebooks,
			FeatureCodeSearch,
		},
	},
}
