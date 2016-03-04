package githubcli

import (
	"net/url"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags defines settings (in the form of CLI flags) for federation.
type GitHubFlags struct {
	// GitHubHost is the hostname of the GitHub instance to mirror repos
	// from. This can point to a GitHub Enterprise instance.
	// NOTE: SSL must be enabled on the GHE instance for mirroring to work.
	GitHubHost string `long:"github.host" description:"hostname of the GitHub (Enterprise) instance to mirror repos from." default:"github.com"`
}

// IsGitHubEnterprise returns true if a GitHub Enterprise instance was
// configured on the CLI.
func (g *GitHubFlags) IsGitHubEnterprise() bool {
	return g.GitHubHost != "github.com"
}

// Host returns the host name of the configured remote GitHub instance.
// eg "github.com" or "ghe.mycompany.com"
func (g *GitHubFlags) Host() string {
	return g.GitHubHost
}

// URL returns the base URL of the configured remote GitHub instance.
// eg "https://github.com" or "https://ghe.mycompany.com"
func (g *GitHubFlags) URL() string {
	return "https://" + g.GitHubHost
}

// APIBaseURL returns the API endpoint URL of the configured GitHub Enterprise instance.
// eg "https://ghe.mycompany.com/api/v3"
func (g *GitHubFlags) APIBaseURL() *url.URL {
	u, err := url.Parse("https://" + g.GitHubHost + "/api/v3/")
	if err != nil {
		return nil
	}
	return u
}

// UploadURL returns the upload endpoint URL of the configured GitHub Enterprise instance.
// eg or "https://ghe.mycompany.com/uploads
func (g *GitHubFlags) UploadURL() *url.URL {
	u, err := url.Parse("https://" + g.GitHubHost + "/uploads")
	if err != nil {
		return nil
	}
	return u
}

// Config is the currently active GitHub config (as set by the CLI flags).
var Config GitHubFlags

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("GitHub", "GitHub", &Config)
	})
}
