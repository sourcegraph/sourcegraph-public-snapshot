package googlesecretsmanager

// ProjectID is the Google Cloud Project that must have the secrets listed in
// this package available in Google Secrets Manager for MSP.
const ProjectID = "sourcegraph-secrets"

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
	SecretOpsgenieAPIToken   = "MSP_OPSGENIE_API_TOKEN"
	SecretSlackOAuthToken    = "SLACK_BOT_USER_OAUTH_TOKEN"

	SecretSourcegraphWildcardKey  = "SOURCEGRAPH_WILDCARD_KEY"
	SecretSourcegraphWildcardCert = "SOURCEGRAPH_WILDCARD_CERT"
)
