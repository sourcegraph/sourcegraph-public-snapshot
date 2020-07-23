package licensing

// The list of plans.
const (
	// oldEnterpriseStarter is the old "Enterprise Starter" plan.
	oldEnterpriseStarter Plan = "old-starter-0"

	// oldEnterprise is the old "Enterprise" plan.
	oldEnterprise Plan = "old-enterprise-0"

	// team is the "Team" plan.
	team Plan = "team-0"
)

var allPlans = []Plan{
	oldEnterpriseStarter,
	oldEnterprise,
	team,
}

// The list of features. For each feature, add a new const here and the checking logic in
// isFeatureEnabled.
const (
	// FeatureACLs is whether ACLs may be used, such as GitHub or GitLab repository permissions and
	// integration with GitHub/GitLab for user authentication.
	FeatureACLs Feature = "acls"

	// FeatureExtensionRegistry is whether publishing extensions to this Sourcegraph instance is
	// allowed. If not, then extensions must be published to Sourcegraph.com. All instances may use
	// extensions published to Sourcegraph.com.
	FeatureExtensionRegistry Feature = "private-extension-registry"

	// FeatureRemoteExtensionsAllowDisallow is whether the site admin may explicitly specify a list
	// of allowed remote extensions and prevent any other remote extensions from being used. It does
	// not apply to locally published extensions.
	FeatureRemoteExtensionsAllowDisallow = "remote-extensions-allow-disallow"
)

// planFeatures defines the features that are enabled for each plan.
var planFeatures = map[Plan][]Feature{
	oldEnterpriseStarter: {},
	oldEnterprise: {
		FeatureACLs,
		FeatureExtensionRegistry,
		FeatureRemoteExtensionsAllowDisallow,
	},
	team: {},
}
