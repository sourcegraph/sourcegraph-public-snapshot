// Package config contains the configuration definition for Sourcegraph Server. This should mirror the configuration
// definition that's used in the editor. The sole responsibility of this package is to provide that definition and
// therefore it should avoid any business logic or third-party dependencies.
package config

import "fmt"

// Config is the app-level configuration for Sourcegraph Server.
// The external README generator (which lives in the infrastructre repository) uses this struct to
// list configuration parameters, using the `json` and `description` tags.
// Fields can be tagged `readme:"enterprise"` to be included only in the Sourcegraph Enterprise README.
//
// Only add configuration fields here that should be visible in config.json to end-user instance admins.
// Internal-only parameters should continue using env.Get.
type Config struct {
	AppID  string `json:"appID" legacyenv:"TRACKING_APP_ID" description:"Application ID to attribute front end user logs to. Providing this value will send usage data back to Sourcegraph (no private code is sent and URLs are sanitized to prevent leakage of private data)."`
	AppURL string `json:"appURL" legacyenv:"SRC_APP_URL" description:"Publicly accessible URL to web app (e.g., what you type into your browser)."`

	DisableTelemetry bool `json:"disableTelemetry" legacyenv:"DISABLE_TELEMETRY" description:"Prevent usage data from being sent back to Sourcegraph (no private code is sent and URLs are sanitized to prevent leakage of private data)."`

	TLSCert string `json:"tlsCert" legacyenv:"TLS_CERT" description:"TLS certificate for the web app."`
	TLSKey  string `json:"tlsKey" legacyenv:"TLS_KEY" description:"TLS key for the web app."`

	CorsOrigin  string `json:"corsOrigin" legacyenv:"CORS_ORIGIN" description:"Value for the Access-Control-Allow-Origin header returned with all requests."`
	AutoRepoAdd bool   `json:"autoRepoAdd" legacyenv:"AUTO_REPO_ADD" description:"Automatically add external public repositories on demand when visited."`

	DisablePublicRepoRedirects bool `json:"disablePublicRepoRedirects" description:"Disable redirects to sourcegraph.com when visiting public repositories that can't exist on this server."`

	Phabricator    []PhabricatorConfig `json:"phabricator" legacyenv:"PHABRICATOR_CONFIG" description:"JSON array of configuration for Phabricator hosts. See Phabricator Configuration section for more information."`
	PhabricatorURL string              `json:"phabricatorURL" legacyenv:"PHABRICATOR_URL" description:"(Deprecated: Use Phabricator) URL of Phabricator instance."`

	GitHub                      []GitHubConfig   `json:"github" legacyenv:"GITHUB_CONFIG" description:"JSON array of configuration for GitHub hosts. See GitHub Configuration section for more information."`
	GithubClientID              string           `json:"githubClientID" legacyenv:"GITHUB_CLIENT_ID" description:"Client ID for GitHub."`
	GithubClientSecret          string           `json:"githubClientSecret" legacyenv:"GITHUB_CLIENT_SECRET" description:"Client secret for GitHub."`
	GithubPersonalAccessToken   string           `json:"githubPersonalAccessToken" legacyenv:"GITHUB_PERSONAL_ACCESS_TOKEN" description:"(Deprecated: Use GitHub) Personal access token for GitHub. "`
	GithubEnterpriseURL         string           `json:"githubEnterpriseURL" legacyenv:"GITHUB_ENTERPRISE_URL" description:"(Deprecated: Use GitHub) URL of GitHub Enterprise instance from which to sync repositories."`
	GithubEnterpriseCert        string           `json:"githubEnterpriseCert" legacyenv:"GITUB_ENTERPRISE_CERT" description:"(Deprecated: Use GitHub) TLS certificate of GitHub Enterprise instance, if from a CA that's not part of the standard certificate chain."`
	GithubEnterpriseAccessToken string           `json:"githubEnterpriseAccessToken" legacyenv:"GITHUB_ENTERPRISE_TOKEN" description:"(Deprecated: Use GitHub) Access token to authenticate to GitHub Enterprise API."`
	GitoliteHosts               string           `json:"gitoliteHosts" legacyenv:"GITOLITE_HOSTS" description:"Space separated list of mappings from repo name prefix to gitolite hosts."`
	GitOriginMap                string           `json:"gitOriginMap" legacyenv:"ORIGIN_MAP" description:"Space separated list of mappings from repo name prefix to origin url, for example \"github.com/!https://github.com/%.git\"."`
	ReposList                   []RepoListConfig `json:"repos.list" legacyenv:"REPOS_LIST" description:"JSON array of configuration for external repositories."`

	InactiveRepos          string `json:"inactiveRepos" legacyenv:"INACTIVE_REPOS" description:"Comma-separated list of repos to consider 'inactive' (e.g. while searching)."`
	LightstepAccessToken   string `json:"lightstepAccessToken" legacyenv:"LIGHTSTEP_ACCESS_TOKEN" description:"Access token for sending traces to LightStep."`
	LightstepProject       string `json:"lightstepProject" legacyenv:"LIGHTSTEP_PROJECT" description:"The project id on LightStep, only used for creating links to traces."`
	NoGoGetDomains         string `json:"noGoGetDomains" legacyenv:"NO_GO_GET_DOMAINS" description:"List of domains to NOT perform go get on. Separated by ','."`
	RepoListUpdateInterval int    `json:"repoListUpdateInterval" legacyenv:"REPO_LIST_UPDATE_INTERVAL" description:"Interval (in minutes) for checking code hosts (e.g. gitolite) for new repositories."`

	SSOUserHeader string `json:"ssoUserHeader" legacyenv:"SSO_USER_HEADER" description:"Header injected by an SSO proxy to indicate the logged in user."`

	OIDCProvider      string `json:"oidcProvider" legacyenv:"OIDC_OP" description:"The URL of the OpenID Connect Provider"`
	OIDCClientID      string `json:"oidcClientID" legacyenv:"OIDC_CLIENT_ID" description:"OIDC Client ID"`
	OIDCClientSecret  string `json:"oidcClientSecret" legacyenv:"OIDC_CLIENT_SECRET" description:"OIDC Client Secret"`
	OIDCEmailDomain   string `json:"oidcEmailDomain" legacyenv:"OIDC_EMAIL_DOMAIN" description:"Whitelisted email domain for logins, e.g. 'mycompany.com'"`
	OIDCOverrideToken string `json:"oidcOverrideToken" legacyenv:"OIDC_OVERRIDE_TOKEN" description:"Token to circumvent OIDC layer (testing only)"`

	SAMLIDProviderMetadataURL string `json:"samlIDProviderMetadataURL" legacyenv:"SAML_ID_PROVIDER_METADATA_URL" description:"SAML Identity Provider metadata URL (for dyanmic configuration of SAML Service Provider)"`
	SAMLSPCert                string `json:"samlSPCert" legacyenv:"SAML_CERT" description:"SAML Service Provider certificate"`
	SAMLSPKey                 string `json:"samlSPKey" legacyenv:"SAML_KEY" description:"SAML Service Provider private key"`

	SearchScopes []SearchScope `json:"searchScopes" legacyenv:"SEARCH_SCOPES" description:"JSON array of custom search scopes (e.g., [{\"name\":\"Text Files\",\"value\":\"file:\\.txt$\"}])"`

	HTMLHeadTop    string `json:"htmlHeadTop" legacyenv:"HTML_HEAD_TOP" description:"HTML to inject at the top of the <head> element on each page, for analytics scripts"`
	HTMLHeadBottom string `json:"htmlHeadBottom" legacyenv:"HTML_HEAD_BOTTOM" description:"HTML to inject at the bottom of the <head> element on each page, for analytics scripts"`
	HTMLBodyTop    string `json:"htmlBodyTop" legacyenv:"HTML_BODY_TOP" description:"HTML to inject at the top of the <body> element on each page, for analytics scripts"`
	HTMLBodyBottom string `json:"htmlBodyBottom" legacyenv:"HTML_BODY_BOTTOM" description:"HTML to inject at the bottom of the <body> element on each page, for analytics scripts"`

	LicenseKey       string `json:"licenseKey" legacyenv:"LICENSE_KEY" description:"License key. You must purchase a license to obtain this."`
	MandrillKey      string `json:"mandrillKey" legacyenv:"MANDRILL_KEY" description:"Key for sending mails via Mandrill."`
	MaxReposToSearch string `json:"maxReposToSearch" legacyenv:"MAX_REPOS_TO_SEARCH" description:"The maximum number of repos to search across. The user is prompted to narrow their query if exceeded. The value 0 means unlimited."`
	AdminUsernames   string `json:"adminUsernames" legacyenv:"ADMIN_USERNAMES" description:"Space-separated list of usernames that indicates which users will be treated as instance admins"`

	Auth Auth `json:"auth"`

	// The following fields are not actually read in this repository, but are used in xlang-java

	ExecuteGradleOriginalRootPaths string `json:"executeGradleOriginalRootPaths" legacyenv:"EXECUTE_GRADLE_ORIGINAL_ROOT_PATHS" description:"Java: A comma-delimited list of patterns that selects repository revisions for which to execute Gradle scripts, rather than extracting Gradle metadata statically. **Security note:** these should be restricted to repositories within your own organization. A percent sign ('%') can be used to prefix-match. For example, <tt>git://my.internal.host/org1/%,git://my.internal.host/org2/repoA?%</tt> would select all revisions of all repositories in org1 and all revisions of repoA in org2."`
	PrivateArtifactRepoID          string `json:"privateArtifactRepoID" legacyenv:"PRIVATE_ARTIFACT_REPO_ID" description:"Java: Private artifact repository ID in your build files. If you do not explicitly include the private artifact repository, then set this to some unique string (e.g,. \"my-repository\")."`
	PrivateArtifactRepoURL         string `json:"privateArtifactRepoURL" legacyenv:"PRIVATE_ARTIFACT_REPO_URL" description:"Java: The URL that corresponds to privateArtifactRepoID (e.g., http://my.artifactory.local/artifactory/root)."`
	PrivateArtifactRepoUsername    string `json:"privateArtifactRepoUsername" legacyenv:"PRIVATE_ARTIFACT_REPO_USERNAME" description:"Java: The username to authenticate to the private Artifactory."`
	PrivateArtifactRepoPassword    string `json:"privateArtifactRepoPassword" legacyenv:"PRIVATE_ARTIFACT_REPO_PASSWORD" description:"Java: The password to authenticate to the private Artifactory."`
}

type Auth struct {
	UserOrgMap UserOrgMap `json:"userOrgMap" description:"Ensure that matching users are members of the specified orgs (auto-joining users to the orgs if they are not already a member). Provide a JSON object of the form <tt>{\"*\": [\"org1\", \"org2\"]}</tt>, where org1 and org2 are orgs that all users are automatically joined to. Currently the only supported key is <tt>\"*\"</tt>."`
}

type GitHubConfig struct {
	URL         string   `json:"url" description:"URL of a GitHub instance e.g. https://github.com."`
	Token       string   `json:"token" description:"A GitHub personal access token with repo and org scope."`
	Certificate string   `json:"certificate,omitempty" description:"TLS certificate of a GitHub Enterprise instance, if from a CA that's not part of the standard certificate chain."`
	Repos       []string `json:"repos,omitempty" description:"Optional whitelist of additional public repos to clone."`
}

type PhabricatorConfig struct {
	URL   string `json:"url" description:"URL of a Phabricator instance, e.g. http://phabricator.mycompany.com."`
	Token string `json:"token" description:"API token for Phabricator instance."`
}

type RepoListConfig struct {
	Type string `json:"type" description:"Type of the version control system for this URL e.g. git"`
	URL  string `json:"url" description:"Clone URL for the repository e.g. git@gitolite.example.com/my/repo.git"`
	Path string `json:"path" description:"Display path for the url e.g. gitolite/my/repo"`
}

type SearchScope struct {
	Name  string `json:"name" description:"User-visible name of search scope"`
	Value string `json:"value" description:"Search scope filter"`
}

// UserOrgMap is a map from user pattern to a list of org names.
type UserOrgMap map[string][]string

// OrgsForAllUsersToJoin returns the list of org names that all users should be joined to. The second return value
// is a list of errors encountered while generating this list. Note that even if errors are returned, the first
// return value is still valid.
func (m UserOrgMap) OrgsForAllUsersToJoin() ([]string, []error) {
	var errors []error
	for userPattern, orgs := range m {
		if userPattern != "*" {
			errors = append(errors, fmt.Errorf("unsupported auth.userOrgMap user pattern %q (only \"*\" is supported)", userPattern))
			continue
		}
		return orgs, errors
	}
	return nil, errors
}
