package conf

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/atomicvalue"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// SearchTimeoutParameterEnabled returns true if SearchTimeoutParameter experiment is enabled.
func SearchTimeoutParameterEnabled() bool {
	p := Get().ExperimentalFeatures.SearchTimeoutParameter
	return p == "enabled"
}

// FileHistorySidebarEnabled returns true if FileHistorySidebar experiment is enabled.
func FileHistorySidebarEnabled() bool {
	p := Get().ExperimentalFeatures.FileHistorySidebar
	return p == "enabled"
}

// EmailVerificationRequired returns whether users must verify an email address before they
// can perform most actions on this site.
//
// It's false for sites that do not have an email sending API key set up.
func EmailVerificationRequired() bool {
	return Get().EmailSmtp != nil
}

// CanSendEmail returns whether the site can send emails (e.g., to reset a password or
// invite a user to an org).
//
// It's false for sites that do not have an email sending API key set up.
func CanSendEmail() bool {
	return Get().EmailSmtp != nil
}

// HasGitHubDotComToken reports whether there are any personal access tokens configured for
// github.com.
func HasGitHubDotComToken() bool {
	for _, c := range Get().Github {
		u, err := url.Parse(c.Url)
		if err != nil {
			continue
		}
		hostname := strings.ToLower(u.Hostname())
		if (hostname == "github.com" || hostname == "api.github.com") && c.Token != "" {
			return true
		}
	}
	return false
}

// HasGitLabDotComToken reports whether there are any personal access tokens configured for
// github.com.
func HasGitLabDotComToken() bool {
	for _, c := range Get().Gitlab {
		u, err := url.Parse(c.Url)
		if err != nil {
			continue
		}
		hostname := strings.ToLower(u.Hostname())
		if hostname == "gitlab.com" && c.Token != "" {
			return true
		}
	}
	return false
}

// EnabledLangservers returns the langservers that are not disabled.
func EnabledLangservers() []schema.Langservers {
	all := Get().Langservers
	results := make([]schema.Langservers, 0, len(all))
	for _, langserver := range all {
		if langserver.Disabled {
			continue
		}
		results = append(results, langserver)
	}
	return results
}

// GitHubEnterpriseURLs is a map of GitHub Enterprise hosts to their full URLs.
// This can be used for the purposes of generating external GitHub enterprise links.
func GitHubEnterpriseURLs() map[string]string {
	return gitHubEnterpriseURLs.Get().(map[string]string)
}

var gitHubEnterpriseURLs = atomicvalue.New()

func addWatchers() {
	Watch(func() {
		gitHubEnterpriseURLs.Set(func() interface{} {
			urls := make(map[string]string)
			for _, c := range Get().Github {
				gheURL, err := url.Parse(c.Url)
				if err != nil {
					log15.Error("error parsing GitHub config", "error", err)
				}
				urls[gheURL.Host] = strings.TrimSuffix(c.Url, "/")
			}
			return urls
		})
	})
}

// DeployType tells the deployment type.
func DeployType() string {
	return os.Getenv("DEPLOY_TYPE")
}

// IsDataCenter tells if the given deployment type is Data Center or, if not, Server.
func IsDataCenter(deployType string) bool {
	return deployType == "datacenter"
}

// DebugManageDocker tells if Docker language servers should be managed or not.
//
// This only exists for dev mode / debugging purposes, and should never be used
// in a production setting.
func DebugManageDocker() bool {
	v, err := strconv.ParseBool(os.Getenv("DEBUG_MANAGE_DOCKER"))
	if err != nil {
		return true // always manage Docker when not set
	}
	return v
}
