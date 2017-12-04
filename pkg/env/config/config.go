package config

// Config is the app-level configuration for Sourcegraph Server.
// The external README generator (which lives in the infrastructre repository) uses this struct to
// list configuration parameters, using the `json` and `description` tags.
// Fields can be tagged `noreadme:"true"` to be omitted from the README.
type Config struct {
	AppID                          string              `json:"appID" description:"Application ID to attribute front end user logs to. Providing this value will send usage data back to Sourcegraph (no private code is sent and URLs are sanitized to prevent leakage of private data)."`
	AppURL                         string              `json:"appURL" description:"Publicly accessible URL to web app (e.g., what you type into your browser)."`
	TLSCert                        string              `json:"tlsCert" description:"TLS certificate for the web app."`
	TLSKey                         string              `json:"tlsKey" description:"TLS key for the web app."`
	CorsOrigin                     string              `json:"corsOrigin" description:"Value for the Access-Control-Allow-Origin header returned with all requests."`
	AutoRepoAdd                    bool                `json:"autoRepoAdd" description:"Automatically add external public repositories on demand when visited."`
	DisablePublicRepoRedirects     bool                `json:"disablePublicRepoRedirects" description:"Disable redirects to sourcegraph.com when visiting public repositories that can't exist on this server."`
	GitHub                         []GitHubConfig      `json:"github" description:"JSON array of configuration for GitHub hosts. See GitHub Configuration section for more information."`
	Phabricator                    []PhabricatorConfig `json:"phabricator" description:"JSON array of configuration for Phabricator hosts. See Phabricator Configuration section for more information."`
	PhabricatorURL                 string              `json:"phabricatorURL" description:"(Deprecated: Use Phabricator) URL of Phabricator instance."`
	GithubClientID                 string              `json:"githubClientID" description:"Client ID for GitHub."`
	GithubClientSecret             string              `json:"githubClientSecret" description:"Client secret for GitHub."`
	GithubPersonalAccessToken      string              `json:"githubPersonalAccessToken" description:"(Deprecated: Use GitHub) Personal access token for GitHub. "`
	GithubEnterpriseURL            string              `json:"githubEnterpriseURL" description:"(Deprecated: Use GitHub) URL of GitHub Enterprise instance from which to sync repositories."`
	GithubEnterpriseCert           string              `json:"githubEnterpriseCert" description:"(Deprecated: Use GitHub) TLS certificate of GitHub Enterprise instance, if from a CA that's not part of the standard certificate chain."`
	GithubEnterpriseAccessToken    string              `json:"githubEnterpriseAccessToken" description:"(Deprecated: Use GitHub) Access token to authenticate to GitHub Enterprise API."`
	GitoliteHosts                  string              `json:"gitoliteHosts" description:"Space separated list of mappings from repo name prefix to gitolite hosts."`
	GitOriginMap                   string              `json:"gitOriginMap" description:"Space separated list of mappings from repo name prefix to origin url, for example \"github.com/!https://github.com/%.git\"."`
	InactiveRepos                  string              `json:"inactiveRepos" description:"Comma-separated list of repos to consider 'inactive' (e.g. while searching)."`
	LightstepAccessToken           string              `json:"lightstepAccessToken" description:"Access token for sending traces to LightStep."`
	LightstepProject               string              `json:"lightstepProject" description:"The project id on LightStep, only used for creating links to traces."`
	NoGoGetDomains                 string              `json:"noGoGetDomains" description:"List of domains to NOT perform go get on. Separated by ','."`
	RepoListUpdateInterval         int                 `json:"repoListUpdateInterval" description:"Interval (in minutes) for checking code hosts (e.g. gitolite) for new repositories."`
	SSOUserHeader                  string              `json:"ssoUserHeader" description:"Header injected by an SSO proxy to indicate the logged in user."`
	StorageClass                   string              `json:"storageClass" description:"Storage class name to use for Persistent Volume claims. If you set this, you need to ensure a storage class with the same name exists in your cluster."`
	ExecuteGradleOriginalRootPaths string              `json:"executeGradleOriginalRootPaths" description:"Java: A comma-delimited list of patterns that selects repository revisions for which to execute Gradle scripts, rather than extracting Gradle metadata statically. **Security note:** these should be restricted to repositories within your own organization. A percent sign ('%') can be used to prefix-match. For example, <tt>git://my.internal.host/org1/%,git://my.internal.host/org2/repoA?%</tt> would select all revisions of all repositories in org1 and all revisions of repoA in org2."`
	OIDCProvider                   string              `json:"oidcProvider" description:"The URL of the OpenID Connect Provider"`
	OIDCClientID                   string              `json:"oidcClientID" description:"OIDC Client ID"`
	OIDCClientSecret               string              `json:"oidcClientSecret" description:"OIDC Client Secret"`
	OIDCEmailDomain                string              `json:"oidcEmailDomain" description:"Whitelisted email domain for logins, e.g. 'mycompany.com'"`
	OIDCOverrideToken              string              `json:"oidcOverrideToken" description:"Token to circumvent OIDC layer (testing only)"`
	SAMLIDProviderMetadataURL      string              `json:"samlIDProviderMetadataURL" description:"SAML Identity Provider metadata URL (for dyanmic configuration of SAML Service Provider)"`
	SAMLSPCert                     string              `json:"samlSPCert" description:"SAML Service Provider certificate"`
	SAMLSPKey                      string              `json:"samlSPKey" description:"SAML Service Provider private key"`
	SearchScopes                   []SearchScope       `json:"searchScopes" description:"JSON array of custom search scopes (e.g., [{\"name\":\"Text Files\",\"value\":\"file:\\.txt$\"}])"`
	PrivateArtifactRepoID          string              `json:"privateArtifactRepoID" description:"Java: Private artifact repository ID in your build files. If you do not explicitly include the private artifact repository, then set this to some unique string (e.g,. \"my-repository\")."`
	PrivateArtifactRepoURL         string              `json:"privateArtifactRepoURL" description:"Java: The URL that corresponds to privateArtifactRepoID (e.g., http://my.artifactory.local/artifactory/root)."`
	HTMLHeadTop                    string              `json:"htmlHeadTop" description:"HTML to inject at the top of the <head> element on each page, for analytics scripts"`
	HTMLHeadBottom                 string              `json:"htmlHeadBottom" description:"HTML to inject at the bottom of the <head> element on each page, for analytics scripts"`
	HTMLBodyTop                    string              `json:"htmlBodyTop" description:"HTML to inject at the top of the <body> element on each page, for analytics scripts"`
	HTMLBodyBottom                 string              `json:"htmlBodyBottom" description:"HTML to inject at the bottom of the <body> element on each page, for analytics scripts"`
	LicenseKey                     string              `json:"licenseKey" description:"License key. You must purchase a license to obtain this."`
	MaxReposToSearch               string              `json:"maxReposToSearch" description:"The maximum number of repos to search across (the user is prompted to narrow their query if exceeded)"`
	AdminUsernames                 string              `json:"adminUsernames" description:"Space-separated list of usernames that indicates which users will be treated as instance admins"`
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

type SearchScope struct {
	Name  string `json:"name" description:"User-visible name of search scope"`
	Value string `json:"value" description:"Search scope filter"`
}
