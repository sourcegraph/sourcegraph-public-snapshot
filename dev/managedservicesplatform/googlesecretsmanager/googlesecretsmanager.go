package googlesecretsmanager

// ProjectID is the Google Cloud Project that must have the secrets listed in
// this package available in Google Secrets Manager for MSP.
const ProjectID = "sourcegraph-secrets"

const (
	/// SecretTFCAccessToken is used for managing TFC workspaces. It cannot
	// be used for creating runs.
	SecretTFCAccessToken = "TFC_ORGANIZATION_TOKEN"
	// SecretTFCOAuthClientID is used for creating VCS-mode workspaces that sync
	// with GitHub.
	SecretTFCOAuthClientID = "TFC_OAUTH_CLIENT_ID"
	// SecretTFCMSPTeamToken is used for creating runs on MSP TFC workspaces.
	// It cannot be used for creating workspaces.
	SecretTFCMSPTeamToken = "TFC_MSP_TEAM_TOKEN"

	SecretCloudflareAPIToken = "CLOUDFLARE_API_TOKEN"

	SecretSourcegraphWildcardKey  = "SOURCEGRAPH_WILDCARD_KEY"
	SecretSourcegraphWildcardCert = "SOURCEGRAPH_WILDCARD_CERT"
)
