package googlesecretsmanager

// SharedSecretsProjectID is the Google Cloud Project that must have the secrets listed in
// this package available in Google Secrets Manager for MSP.
const SharedSecretsProjectID = "sourcegraph-secrets"

const (
	/// SecretTFCOrgToken is used for managing TFC workspaces. It cannot
	// be used for creating runs. Other than initial workspace creation, prefer
	// to use SecretTFCMSPTeamToken instead.
	SecretTFCOrgToken = "TFC_ORGANIZATION_TOKEN"
	// SecretTFCMSPTeamToken is used for creating runs on MSP TFC workspaces.
	// It cannot be used for creating workspaces. Other than initial workspace
	// creation, for which SecretTFCAccessToken should be used, you should use
	// this token instead.
	SecretTFCMSPTeamToken = "TFC_MSP_TEAM_TOKEN"
	// SecretTFCOAuthClientID is used for creating VCS-mode workspaces that sync
	// with GitHub.
	SecretTFCOAuthClientID = "TFC_OAUTH_CLIENT_ID"
	// SecretTFCMSPSlackWebhook is the Slack webhook for use in TFC notifications
	// from workspaces managed by MSP.
	SecretTFCMSPSlackWebhook = "TFC_MSP_SLACK_WEBHOOK"

	SecretCloudflareAPIToken = "CLOUDFLARE_API_TOKEN"
	// SecretOpsgenieAPIToken is an Opsgenie token with integration management
	// privileges.
	SecretOpsgenieAPIToken = "MSP_OPSGENIE_API_TOKEN"
	// SecretSlackOAuthToken is used for managing Slack notification integrations.
	// We just use the one that seems to be used elsewhere as well.
	SecretSlackOAuthToken = "SLACK_BOT_USER_OAUTH_TOKEN"
	// SecretSlackOperatorOAuthToken is used for managing public Slack channels.
	// It needs to be a bot user token with the scopes documented in
	// https://registry.terraform.io/providers/pablovarela/slack/latest/docs/resources/conversation
	//
	// The current bot user is https://api.slack.com/apps/A06C4TF6YF7/oauth
	SecretSlackOperatorOAuthToken = "SLACK_OPERATOR_BOT_OAUTH_TOKEN"
	// SecretSentryAuthToken is a Sentry internal integration auth token with Project permissions.
	// The integration is configured in https://sourcegraph.sentry.io/settings/developer-settings/managed-services-platform-fbf7cc/
	SecretSentryAuthToken = "TFC_MSP_SENTRY_INTEGRATION"
	// SecretNobl9ClientSecret is used to provision Nobl9 projects
	SecretNobl9ClientSecret = "MSP_NOBL9_CLIENT_SECRET"
	// SecretSourcegraphWildcardKey and SecretSourcegraphWildcardCert are used
	// for configuring Cloudflare TLS.
	SecretSourcegraphWildcardKey  = "SOURCEGRAPH_WILDCARD_KEY"
	SecretSourcegraphWildcardCert = "SOURCEGRAPH_WILDCARD_CERT"
	// SecretMSPDeployNotificationEndpoint is the endpoint that MSP uses for cloud deploy push notifications.
	SecretMSPDeployNotificationEndpoint = "MSP_DEPLOY_NOTIFICATION_ENDPOINT"
)
