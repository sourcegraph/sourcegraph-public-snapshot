package githubcli

import (
	"net/url"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

// Flags defines settings (in the form of CLI flags) related to
// GitHub.
type GitHubFlags struct {
	// GitHubHost is the hostname of the GitHub instance to mirror repos
	// from. This can point to a GitHub Enterprise instance.
	// NOTE: SSL must be enabled on the GHE instance for mirroring to work.
	GitHubHost string `long:"github.host" description:"hostname of the GitHub (Enterprise) instance to mirror repos from." default:"github.com"`
	Disable    bool   `long:"github.disable" description:"disables communication with GitHub instances. Used to test GitHub service degredation" env:"SRC_GITHUB_DISABLE"`
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

var gitHubHost = env.Get("SRC_GITHUB_HOST", "github.com", "hostname of the GitHub (Enterprise) instance to mirror repos from")
var gitHubDisable, _ = strconv.ParseBool(env.Get("SRC_GITHUB_DISABLE", "false", "disables communication with GitHub instances. Used to test GitHub service degredation"))

// Config is the currently active GitHub config (as set by the CLI flags).
var Config = GitHubFlags{
	GitHubHost: gitHubHost,
	Disable:    gitHubDisable,
}
