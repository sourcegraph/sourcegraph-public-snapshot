package googlesecretsmanager

// ProjectID is the Google Cloud Project that must have the secrets listed in
// this package available in Google Secrets Manager for MSP.
const ProjectID = "sourcegraph-secrets"

var (
	SecretGCPBillingAccount  = "MSP_BILLING_ACCOUNT_ID" // TODO: Add
	SecretGCPProjectFolderID = "MSP_PARENT_FOLDER_ID"   // TODO: Add

	SecretTFCAccessToken   = "TFC_ORGANIZATION_TOKEN"
	SecretTFCOAuthClientID = "TFC_OAUTH_CLIENT_ID"

	SecretCloudflareAPIToken = "CLOUDFLARE_API_TOKEN"

	SecretSourcegraphWildcardKey  = "SOURCEGRAPH_WILDCARD_KEY"
	SecretSourcegraphWildcardCert = "SOURCEGRAPH_WILDCARD_CERT"
)
